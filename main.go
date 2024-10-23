package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	"github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func main() {
	opt := &option{}
	cmd := &cobra.Command{
		Use:  "argo-wf-atomic",
		RunE: opt.runE,
	}
	flags := cmd.Flags()
	flags.IntVarP(&opt.port, "port", "", 3002, "The port of the HTTP server")
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func (o *option) runE(c *cobra.Command, args []string) (err error) {
	var config *rest.Config
	if config, err = rest.InClusterConfig(); err != nil {
		return
	}
	client := wfclientset.NewForConfigOrDie(config)

	http.HandleFunc("/api/v1/template.execute", plugin(client))
	err = http.ListenAndServe(fmt.Sprintf(":%d", o.port), nil)
	return
}

type option struct {
	port int
}

var (
	errWrongContentType = errors.New("Content-Type header is not set to 'appliaction/json'")
	errReadingBody      = errors.New("couldn't read request body")
	errMarshallingBody  = errors.New("couldn't unmrashal request body")
)

func plugin(client *wfclientset.Clientset) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var isNot404 bool
		var err error
		defer func() {
			if !isNot404 {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			var nodeResult *wfv1.NodeResult
			if err == nil {
				nodeResult = &wfv1.NodeResult{
					Phase:   wfv1.NodeSucceeded,
					Message: "success",
				}
			} else {
				nodeResult = &wfv1.NodeResult{
					Phase:   wfv1.NodeError,
					Message: err.Error(),
				}
			}

			jsonResp, jsonErr := json.Marshal(executor.ExecuteTemplateReply{
				Node: nodeResult,
			})
			if jsonErr != nil {
				fmt.Println("something went wrong", jsonErr)
				http.Error(w, "something went wrong", http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(jsonResp)
			}
		}()

		if header := req.Header.Get("Content-Type"); header != "application/json" {
			err = errWrongContentType
			return
		}

		var body []byte
		if body, err = io.ReadAll(req.Body); err != nil {
			err = errReadingBody
			return
		}

		fmt.Println(string(body))
		args := executor.ExecuteTemplateArgs{}
		if err = json.Unmarshal(body, &args); err != nil || args.Workflow == nil || args.Template == nil {
			err = errMarshallingBody
			return
		}

		p := args.Template.Plugin.Value
		opt := map[string]interface{}{}
		if err = json.Unmarshal(p, &opt); err != nil {
			return
		}
		if _, isNot404 = opt["argo-atomic-plugin"]; !isNot404 {
			return
		}

		ns := args.Workflow.ObjectMeta.Namespace
		wfName := args.Workflow.ObjectMeta.Name
		ctx := context.Background()

		// find the Workflow
		var workflow *wfv1.Workflow
		if workflow, err = client.ArgoprojV1alpha1().Workflows(ns).Get(
			context.Background(),
			wfName,
			v1.GetOptions{}); err != nil {
			fmt.Println("failed to find workflow", wfName, ns, err)
			return
		}

		if workflow.Spec.WorkflowTemplateRef == nil {
			fmt.Println("not belong to a template: ", wfName, ns)
			return
		}

		var workflows *wfv1.WorkflowList
		if workflows, err = client.ArgoprojV1alpha1().Workflows(ns).List(ctx, v1.ListOptions{
			LabelSelector: "workflows.argoproj.io/phase=Running",
		}); err != nil {
			return
		}

		for _, wf := range workflows.Items {
			if wf.Spec.WorkflowTemplateRef == nil || wf.Spec.WorkflowTemplateRef.Name != workflow.Spec.WorkflowTemplateRef.Name {
				continue
			}
			if wf.Name == wfName {
				continue
			}
			if !reflect.DeepEqual(wf.Spec.Arguments, workflow.Spec.Arguments) {
				continue
			}

			wfCopied := wf.DeepCopy()
			wfCopied.Spec.Shutdown = wfv1.ShutdownStrategyStop
			if _, err = client.ArgoprojV1alpha1().Workflows(ns).Update(ctx, wfCopied, v1.UpdateOptions{}); err != nil {
				return
			}
		}
	}
}
