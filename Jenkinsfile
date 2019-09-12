node {
    def overrides = [
        scriptVersion  : 'v6',
        pipelineScript: 'https://git.aurora.skead.no/scm/ao/aurora-pipeline-scripts.git',
        credentialsId: "github"
    ]

    def tagVersion

    stage ('Load shared libraries') {
           def version='v6'
           fileLoader.withGit(overrides.pipelineScript,, overrides.scriptVersion) {
               maven = fileLoader.load('maven/maven')
               git = fileLoader.load('git/git')
               go = fileLoader.load('go/go')
               openshift = fileLoader.load('openshift/openshift')
           }
       }

    stage ('Checkout') {
        checkout scm
    }

    stage('Test and coverage'){
        go.buildGoWithJenkinsSh()
    }

    stage('Sonar') {
        def sonarPath = tool 'Sonar 4'
        sh "${sonarPath}/bin/sonar-scanner -Dsonar.branch.name=${env.BRANCH_NAME}"
    }

    stage('Deploy to Nexus') {
        def isMaster = env.BRANCH_NAME == 'master'
        tagVersion = git.executeAuroraGitVersionCliCommand(" --suggest-releases master --version-hint 1 --increment-for-existing-tag")

        if (isMaster){
            git.tagIfNotExists('ci_aos', tagVersion)
        }

        maven.deployTarGzToNexusWithGroupId("bin/amd64/", "architect", "ske.aurora.openshift", tagVersion)
    }

    stage ('OpenShift build') {
        def namespace = openshift.jenkinsNamespace()
        def result = openshift.oc("start-build architect -e ARTIFACT_NAME=architect -e GROUP_ID=ske.aurora.openshift -e ARTIFACT_VERSION=${tagVersion} -n=${namespace} -F")
        if(!result) {
            error("Building docker image failed")
        }
    }
}


