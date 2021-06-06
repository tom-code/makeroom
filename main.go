package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"

  admission "k8s.io/api/admission/v1beta1"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


func printJson(data []byte) {
  var out bytes.Buffer
  var js string
  err := json.Indent(&out, []byte(data), "", "  ")
  if err != nil {
    log.Printf("json decode error: %s\n", err.Error())
    js = string(data)
  } else {
    js = out.String()
  }
  log.Println(js)
}


type patchItem struct {
  Op string `json:"op"`
  Path string `json:"path"`
}

func preparePatch(pod corev1.Pod) []byte {
  patch := []patchItem{}
  for i, cont := range pod.Spec.Containers {
    if len(cont.Resources.Requests) != 0 {
      patch = append(patch, patchItem{
        Op: "remove",
        Path: fmt.Sprintf("/spec/containers/%d/resources", i)},
      )
    }
  }
  out, err := json.Marshal(&patch)
  if err != nil {
    log.Println(err)
    return []byte("[]")
  }
  log.Println(string(out))
  return out
}

type Handler struct {}
func (h* Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("new request %s\n", r.RequestURI)
  data, err := ioutil.ReadAll(r.Body)
  r.Body.Close()
  if err != nil {
    log.Println(err)
  } else {
    patch := []byte("[]")
    printJson(data)
    var review admission.AdmissionReview
    err = json.Unmarshal(data, &review)
    if err != nil {
      log.Println(err)
    } else {
      log.Println(review.Request.UID)
      var pod corev1.Pod
      err = json.Unmarshal(review.Request.Object.Raw, &pod)
      if err != nil {
        log.Println(err)
      } else {
        patch = preparePatch(pod)
      }
    }
    patchType := admission.PatchTypeJSONPatch
    resp := admission.AdmissionReview{
      TypeMeta: metav1.TypeMeta {
        Kind: "AdmissionReview",
        APIVersion: "admission.k8s.io/v1" },
      Response: &admission.AdmissionResponse {
        UID: review.Request.UID,
        Allowed: true,
        PatchType: &patchType,
        Patch: patch,
      },
    }
    respb, err := json.Marshal(&resp)
    if err != nil {
      log.Println(err)
    }
    log.Println(string(respb))
    w.Write(respb)
  }
}

func main() {
  log.Println("start")

  server := &http.Server{
    Addr:    ":443",
    Handler: &Handler{},
  }
  err := server.ListenAndServeTLS("/cert.pem", "/private.pem")
  if err != nil {
    panic(err)
  }
}
