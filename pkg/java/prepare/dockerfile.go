package prepare

import (
	"io"
	"text/template"
	"github.com/Skatteetaten/architect/pkg/java/config"
)

type Dockerfile interface {
	Build(writer io.Writer) (error)
}

var dockerfileTemplate string =

	`FROM {{.DockerBase}}

	MAINTAINER {{.Maintainer}}
	LABEL {{range $key, $value := .Labels}}{{$key}}="{{$value}}" {{end}}

	COPY ./app /u01
	RUN chmod -R 777 /u01/

	ENV {{range $key, $value := .Env}}{{$key}}="{{$value}}" {{end}}

	CMD ["bin/run"]`

type DefaultDockerfile struct {
	DockerBase string
	Maintainer string
	Labels     map[string]string
	Env        map[string]string
}

func NewForConfig(DockerBase string, Env map[string]string, cfg *config.ArchitectConfig) Dockerfile {
	var impl *DefaultDockerfile = &DefaultDockerfile{}
	impl.Maintainer = cfg.Docker.Maintainer
	impl.Labels = cfg.Docker.Labels
	impl.Env = Env
	impl.DockerBase = DockerBase
	appendReadinesEnv(impl, cfg)
	var spec Dockerfile = impl
	return spec
}

func (dockerfile *DefaultDockerfile) Build(writer io.Writer) (error) {

	tmpl, err := template.New("dockerfile").Parse(dockerfileTemplate)

	if err != nil {
		return err
	}

	return tmpl.Execute(writer, dockerfile)
}

func appendReadinesEnv(dockerfile *DefaultDockerfile, cfg *config.ArchitectConfig) {

	if cfg.Openshift != nil {
		if cfg.Openshift.ReadinessURL != "" {
			dockerfile.Env["READINESS_CHECK_URL"] = cfg.Openshift.ReadinessURL
		}

		if cfg.Openshift.ReadinessOnManagementPort == "" || cfg.Openshift.ReadinessOnManagementPort == "true" {
			dockerfile.Env["READINESS_ON_MANAGEMENT_PORT"] = "true"
		}
	} else if cfg.Java != nil && cfg.Java.ReadinessURL != "" {
		dockerfile.Env["READINESS_CHECK_URL"] = cfg.Java.ReadinessURL
	}
}
