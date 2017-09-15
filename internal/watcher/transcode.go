package watcher

import (
	"log"
	"path/filepath"

	"github.com/carolynvs/handbrk8s/internal/k8s/jobs"
)

const transcodeJobYaml = `
apiVersion: batch/v1
kind: Job
metadata:
  name: transcode-{{.Name}}
  namespace: handbrk8s
spec:
  template:
    metadata:
      name: handbrake-{{.Name}}
    spec:
      containers:
      - name: handbrake
        image: carolynvs/handbrakecli:latest
        imagePullPolicy: Always
        args:
        - "--preset-import-file"
        - "/config/ghb/presets.json"
        - "-i"
        - "{{.InputPath}}"
        - "-o"
        - "{{.OutputPath}}"
        - "--preset"
        - "{{.Preset}}"
        volumeMounts:
        - mountPath: /work
          name: ponyshare
      restartPolicy: OnFailure
      volumes:
      - name: ponyshare
        hostPath:
          path: /mlp
`

// TranscodeJobValues are the set of values to replace in transcodeJobYaml
type transcodeJobValues struct {
	Name, InputPath, OutputPath, Preset string
}

// CreateTranscodeJob creates a job to transcode a video
func (w *VideoWatcher) createTranscodeJob(inputPath string, outputPath string) (jobName string, err error) {
	filename := filepath.Base(inputPath)

	log.Printf("creating transcode job for %s\n", filename)
	values := transcodeJobValues{
		Name:       jobs.SanitizeJobName(filename),
		InputPath:  inputPath,
		OutputPath: outputPath,
		Preset:     w.VideoPreset,
	}
	return jobs.CreateFromTemplate(transcodeJobYaml, values)
}