package confserver

import (
	"bytes"
	"fmt"
	"path/filepath"
)

func (c *Configuration) dockerfile() []byte {
	dockerfile := bytes.NewBuffer(nil)

	_, _ = fmt.Fprintln(dockerfile, `
FROM docker.io/library/golang:1.17-buster AS build-env

FROM build-env AS builder

WORKDIR /go/src
COPY ./ ./

# build
RUN make build WORKSPACE=`+c.WorkSpace()+`

# runtime
FROM alpine
COPY --from=builder `+filepath.Join("/go/src/cmd", c.WorkSpace(), c.WorkSpace())+` `+filepath.Join(`/go/bin`, c.Command.Use)+`
`)
	for _, envVar := range c.defaultEnvVars.Values {
		if envVar.Value != "" {
			if envVar.IsCopy {
				_, _ = fmt.Fprintf(dockerfile, "COPY --from=builder %s %s\n", filepath.Join("/go/src/cmd", c.WorkSpace(), envVar.Value), filepath.Join("/go/bin", envVar.Value))
			}
			if envVar.IsExpose {
				_, _ = fmt.Fprintf(dockerfile, "EXPOSE %s\n", envVar.Value)
			}
		}
	}

	fmt.Fprintf(dockerfile, `
ARG PROJECT_NAME
ARG PROJECT_VERSION
ENV GOENV=DEV PROJECT_NAME=${PROJECT_NAME} PROJECT_VERSION=${PROJECT_VERSION}

WORKDIR /go/bin
ENTRYPOINT ["`+filepath.Join(`/go/bin`, c.Command.Use)+`"]
`)

	return dockerfile.Bytes()
}
