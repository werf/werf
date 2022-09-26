package helm

import (
	"bytes"
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/werf"
)

type ExtraAnnotationsAndLabelsPostRendererTestData struct {
	PostRenderer          *ExtraAnnotationsAndLabelsPostRenderer
	Manifest              string
	ExpectedManifest      string
	ExpectedWarningsCount int
}

type globalWarningsMock struct {
	messages []string
}

func (gw *globalWarningsMock) GlobalWarningLn(ctx context.Context, msg string) {
	gw.messages = append(gw.messages, msg)
}

var _ = Describe("ExtraAnnotationsAndLabelsPostRenderer", func() {
	DescribeTable("parsing manifests stream and injecting custom annotations and labels",
		func(data ExtraAnnotationsAndLabelsPostRendererTestData) {
			globalWarnings := &globalWarningsMock{}
			data.PostRenderer.globalWarnings = globalWarnings

			out, err := data.PostRenderer.Run(bytes.NewBuffer([]byte(data.Manifest)))
			Expect(err).ShouldNot(HaveOccurred())

			fmt.Printf("OUT:\n%s\n---\n", out.String())
			fmt.Printf("EXPECTED OUT:\n%s\n---\n", data.ExpectedManifest)
			Expect(out.String()).To(Equal(data.ExpectedManifest))
			for _, msg := range globalWarnings.messages {
				fmt.Printf("WARNING: %s\n", msg)
			}
			Expect(len(globalWarnings.messages)).To(Equal(data.ExpectedWarningsCount))
		},

		Entry("should add builtin extra annotations into resources manifests",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(nil, nil, false),
				Manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
  annotations:
    one: two
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: result
  name: result
spec:
  ports:
    - name: "result-service"
      port: 80
  selector:
    app: result
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
  annotations:
    one: two
    werf.io/version: %s
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: result
  name: result
  annotations:
    werf.io/version: %s
spec:
  ports:
    - name: "result-service"
      port: 80
  selector:
    app: result
`, werf.Version, werf.Version),
			}),

		Entry("should add builtin and extra annotations and labels into resources manifests",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(
					map[string]string{"test-annotation-1": "value-1", "test-annotation-2": "value-2"},
					map[string]string{"test-label-1": "value-1", "test-label-2": "value-2"},
					false,
				),
				Manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
  annotations:
    one: two
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: result
  name: result
spec:
  ports:
    - name: "result-service"
      port: 80
  selector:
    app: result
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
    test-label-1: value-1
    test-label-2: value-2
  annotations:
    one: two
    test-annotation-1: value-1
    test-annotation-2: value-2
    werf.io/version: %s
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: result
    test-label-1: value-1
    test-label-2: value-2
  name: result
  annotations:
    test-annotation-1: value-1
    test-annotation-2: value-2
    werf.io/version: %s
spec:
  ports:
    - name: "result-service"
      port: 80
  selector:
    app: result
`, werf.Version, werf.Version),
			}),

		Entry("should add builtin and extra annotations and labels into resources manifest with empty annotations and labels",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(
					map[string]string{"test-annotation-1": "value-1", "test-annotation-2": "value-2"},
					map[string]string{"test-label-1": "value-1", "test-label-2": "value-2"},
					false,
				),
				Manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
  annotations:
  name: test
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    test-label-1: value-1
    test-label-2: value-2
  annotations:
    test-annotation-1: value-1
    test-annotation-2: value-2
    werf.io/version: %s
  name: test
`, werf.Version),
			}),

		Entry("should add extra annotations into yaml alias node defined by yaml anchor",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(
					map[string]string{"test-annotation-1": "value-1", "test-annotation-2": "value-2"},
					nil,
					false,
				),
				Manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels: &named_anchor
    app: vote
  annotations: *named_anchor
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels: &named_anchor
    app: vote
  annotations: {app: vote, test-annotation-1: value-1, test-annotation-2: value-2, werf.io/version: %s}
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`, werf.Version),
			}),

		Entry("should add extra annotations into yaml node which references some yaml anchor",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(
					map[string]string{"test-annotation-1": "value-1", "test-annotation-2": "value-2"},
					nil,
					false,
				),
				Manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels: &named_anchor
    app: vote
  annotations:
    <<: *named_anchor
    test-key: test-value
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels: &named_anchor
    app: vote
  annotations:
    !!merge <<: *named_anchor
    test-key: test-value
    test-annotation-1: value-1
    test-annotation-2: value-2
    werf.io/version: %s
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`, werf.Version),
			}),

		Entry("should print warnings for invalid annotations and labels which are not strings, should render invalid values though",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(
					nil,
					nil,
					false,
				),
				ExpectedWarningsCount: 2,
				Manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
    one: 1
  annotations:
    hello: world
    two: 2
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
    one: 1
  annotations:
    hello: world
    two: 2
    werf.io/version: %s
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`, werf.Version),
			}),

		Entry("should print warnings for invalid annotations and labels which are not strings, should omit invalid values",
			ExtraAnnotationsAndLabelsPostRendererTestData{
				PostRenderer: NewExtraAnnotationsAndLabelsPostRenderer(
					nil,
					nil,
					true,
				),
				ExpectedWarningsCount: 2,
				Manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
    one: 1
  annotations:
    hello: world
    two: 2
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`,
				ExpectedManifest: fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vote
  annotations:
    hello: world
    werf.io/version: %s
  name: vote
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vote
  template:
    metadata:
      labels:
        app: vote
    spec:
      imagePullSecrets:
        - name: registrysecret
      containers:
        - image: myimage
          name: vote
          ports:
            - containerPort: 80
              name: vote
`, werf.Version),
			}),
	)
})
