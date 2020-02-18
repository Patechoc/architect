package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

type KubernetesConfigReader struct {
	Extractor Extractor
}

type EnvExtractor struct {
}

type Extractor interface {
	ApplicationType() (ApplicationType)
	ApplicationSpec(applicationType ApplicationType) (ApplicationSpec, error)
	BaseImageSpec() (DockerBaseImageSpec, error)
	DockerSpec() (DockerSpec, error)
	NexusAccess() (NexusAccess, error)
	BuilderSpec() (BuilderSpec, error)
}

func NewKubernetesConfigReader(extractor Extractor) ConfigReader {
	return &KubernetesConfigReader{
		Extractor: extractor,
	}
}

func (m *KubernetesConfigReader) ReadConfig() (*Config, error) {

	applicationType := m.Extractor.ApplicationType()
	applicationSpec, err := m.Extractor.ApplicationSpec(applicationType)
	if err != nil {
		return nil, err
	}

	dockerSpec, err := m.Extractor.DockerSpec()
	if err != nil {
		return nil, err
	}

	nexusAccess, err := m.Extractor.NexusAccess()
	if err != nil {
		return nil, err
	}

	builderSpec, err := m.Extractor.BuilderSpec()
	if err != nil {
		return nil, err
	}

	return &Config{
		BuildStrategy:   Buildah,
		BuildTimeout:    900,
		TlsVerify:       false,
		ApplicationType: JavaLeveransepakke,
		ApplicationSpec: applicationSpec,
		DockerSpec:      dockerSpec,
		BuilderSpec:     builderSpec,
		NexusAccess:     nexusAccess,
	}, nil
}

func (e EnvExtractor) ApplicationType() (ApplicationType) {

	var applicationType = JavaLeveransepakke
	if appType, err := lookupEnv("APPLICATION_TYPE"); err == nil {
		if strings.ToUpper(appType) == NodeJs {
			applicationType = NodeJsLeveransepakke
		} else if strings.ToUpper(appType) == Doozer {
			applicationType = DoozerLeveranse
		}
	}
	return applicationType

}

func (e EnvExtractor) ApplicationSpec(applicationType ApplicationType) (ApplicationSpec, error) {

	applicationSpec := ApplicationSpec{}
	if artifactId, err := lookupEnv("ARTIFACT_ID"); err == nil {
		applicationSpec.MavenGav.ArtifactId = artifactId
	} else {
		return ApplicationSpec{}, err
	}
	if groupId, err := lookupEnv("GROUP_ID"); err == nil {
		applicationSpec.MavenGav.GroupId = groupId
	} else {
		return ApplicationSpec{}, err
	}
	if version, err := lookupEnv("VERSION"); err == nil {
		applicationSpec.MavenGav.Version = version
	} else {
		return ApplicationSpec{}, err
	}
	if classifier, err := lookupEnv("CLASSIFIER"); err == nil {
		applicationSpec.MavenGav.Classifier = Classifier(classifier)
	} else {
		if applicationType == JavaLeveransepakke {
			applicationSpec.MavenGav.Classifier = Leveransepakke
		} else if applicationType == NodeJsLeveransepakke {
			applicationSpec.MavenGav.Classifier = Webleveransepakke
		} else {
			applicationSpec.MavenGav.Classifier = Doozerleveransepakke
		}
	}
	if applicationType == JavaLeveransepakke || applicationType == DoozerLeveranse {
		applicationSpec.MavenGav.Type = ZipPackaging
	} else {
		applicationSpec.MavenGav.Type = TgzPackaging
	}

	if baseSpec, err := e.BaseImageSpec(); err == nil {
		applicationSpec.BaseImageSpec = baseSpec
	} else {
		return ApplicationSpec{}, err
	}

	return applicationSpec, nil
}

func (e EnvExtractor) DockerSpec() (DockerSpec, error) {

	dockerSpec := DockerSpec{}

	if externalRegistry, err := lookupEnv("BASE_IMAGE_REGISTRY"); err == nil {
		if strings.HasPrefix(externalRegistry, "https://") {
			dockerSpec.ExternalDockerRegistry = externalRegistry
		} else {
			dockerSpec.ExternalDockerRegistry = "https://" + externalRegistry
		}
	} else {
		dockerSpec.ExternalDockerRegistry = FallbackDockerRegistry
		logrus.Warnf("Failed to find a specified url for ExternalDockerRegistry. Using %s", FallbackDockerRegistry)
	}

	if internalPullRegistry, err := lookupEnv("INTERNAL_PULL_REGISTRY"); err == nil {

		if strings.HasPrefix(internalPullRegistry, "https://") {
			dockerSpec.InternalPullRegistry = internalPullRegistry
		} else {
			dockerSpec.InternalPullRegistry = "https://" + internalPullRegistry
		}
	} else {
		dockerSpec.InternalPullRegistry = FallbackDockerRegistry
		logrus.Warnf("Failed to find a specified url for InternalPullRegistry. Using %s", FallbackDockerRegistry)
	}

	if pushExtraTags, err := lookupEnv("PUSH_EXTRA_TAGS"); err == nil {
		dockerSpec.PushExtraTags = ParseExtraTags(pushExtraTags)
	} else {
		dockerSpec.PushExtraTags = ParseExtraTags("latest,major,minor,patch")
	}

	if temporaryTag, err := lookupEnv("TAG_WITH"); err == nil {
		dockerSpec.TagWith = temporaryTag
	}

	if temporaryTag, err := lookupEnv("RETAG_WITH"); err == nil {
		dockerSpec.RetagWith = temporaryTag
	}

	dockerSpec.TagOverwrite = false
	if tagOverwrite, err := lookupEnv("TAG_OVERWRITE"); err == nil {
		if strings.Contains(strings.ToLower(tagOverwrite), "true") {
			dockerSpec.TagOverwrite = true
		}
	}

	if outputRegistry, err := lookupEnv("OUTPUT_REGISTRY"); err == nil {
		dockerSpec.OutputRegistry = outputRegistry
	}

	if outputRepository, err := lookupEnv("OUTPUT_REPOSITORY"); err == nil {
		dockerSpec.OutputRepository = outputRepository
	}

	return dockerSpec, nil
}

func (e EnvExtractor) NexusAccess() (NexusAccess, error) {

	nexusAccess := NexusAccess{}
	secretPath := "/u01/nexus/nexus.json"
	jsonFile, err := ioutil.ReadFile(secretPath)
	if err == nil {
		var data map[string]interface{}
		err := json.Unmarshal(jsonFile, &data)
		if err != nil {
			return NexusAccess{}, errors.Wrapf(err, "Could not parse %s. Must be correct json when specified.", secretPath)
		}
		nexusAccess.NexusUrl = data["nexusUrl"].(string)
		nexusAccess.Username = data["username"].(string)
		nexusAccess.Password = data["password"].(string)
	} else {
		logrus.Warnf("Could not read nexus config at %s, error: %s", secretPath, err)
	}

	return nexusAccess, nil
}

func (e EnvExtractor) BuilderSpec() (BuilderSpec, error) {

	builderSpec := BuilderSpec{}

	if builderVersion, present := os.LookupEnv("APP_VERSION"); present {
		builderSpec.Version = builderVersion
	} else {
		//We set it to local for local builds.
		//Running on OpenShift will have APP_VERSION as environment variable
		builderSpec.Version = "local"
	}

	return builderSpec, nil
}

func (e EnvExtractor) BaseImageSpec() (DockerBaseImageSpec, error) {
	baseSpec := DockerBaseImageSpec{}
	if baseImage, err := lookupEnv("DOCKER_BASE_IMAGE"); err == nil {
		baseSpec.BaseImage = baseImage
	} else if baseImage, err := lookupEnv("DOCKER_BASE_NAME"); err == nil {
		baseSpec.BaseImage = baseImage
	} else {
		return baseSpec, err
	}

	if baseImageVersion, err := lookupEnv("DOCKER_BASE_VERSION"); err == nil {
		baseSpec.BaseVersion = baseImageVersion
	} else {
		return baseSpec, err
	}
	return baseSpec, nil
}

func lookupEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if ok {
		return value, nil
	}
	return "", errors.New("No env variable with name " + name)
}
