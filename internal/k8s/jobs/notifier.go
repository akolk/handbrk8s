package jobs

import (
	"log"

	"github.com/carolynvs/handbrk8s/internal/k8s/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	k8sapi "k8s.io/client-go/pkg/api"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

func WaitUntilComplete(done <-chan struct{}, namespace, name string) (<-chan *batchv1.Job, <-chan error) {
	jobChan := make(chan *batchv1.Job)
	errChan := make(chan error)

	go func() {
		defer close(jobChan)
		defer close(errChan)

		clusterClient, err := api.GetCurrentClusterClient()
		if err != nil {
			errChan <- err
			return
		}

		jobclient := clusterClient.BatchV1Client.Jobs(namespace)

		opts := metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(k8sapi.ObjectNameField, name).String(),
		}
		watch, err := jobclient.Watch(opts)
		if err != nil {
			errChan <- errors.Wrapf(err, "Unable to watch %v:jobs for %#v", namespace, opts)
			return
		}
		defer watch.Stop()
		events := watch.ResultChan()

		for {
			select {
			case <-done:
				return
			case e := <-events:
				job, ok := e.Object.(*batchv1.Job)
				if !ok {
					errChan <- errors.Errorf("%s", e.Object)
					continue
				}
				if job.Status.Succeeded > 0 {
					jobChan <- job
				} else {
					log.Printf("%#v", job.Status)
				}
			}
		}
	}()

	return jobChan, errChan
}
