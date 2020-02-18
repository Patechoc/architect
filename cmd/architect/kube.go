package architect

import (
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/architect/pkg/config"
	"github.com/skatteetaten/architect/pkg/docker"
	"github.com/skatteetaten/architect/pkg/nexus"
	"github.com/spf13/cobra"
)


func init() {
	Kube.Flags().StringP("file", "f", "", "Path to a kubernetes job defintion")
}


var Kube = &cobra.Command{
	Use: "kube",
	Short: "Build images on kubernetes",
	Run: func(cmd *cobra.Command, args []string) {

		logrus.Info("Run architect on kubernetes")
		var configReader = config.NewKubernetesConfigReader(config.EnvExtractor{})
		var nexusDownloader nexus.Downloader

		config, err := configReader.ReadConfig()
		if err != nil {
			logrus.Fatalf("Unable to read kubernetes build config: %s", err)
			return
		}

		nexusDownloader = nexus.NewNexusDownloader(config.NexusAccess.NexusUrl)

		RunArchitect(RunConfiguration{
			NexusDownloader:nexusDownloader,
			Config: config,
			//TODO: Denne må mest sannsynlig skrives om
			RegistryCredentialsFunc: docker.CusterRegistryCredentials(),
		})

	},

}
