@Library('vega-shared-library') _

/* properties of scmVars (example):
    - GIT_BRANCH:PR-40-head
    - GIT_COMMIT:05a1c6fbe7d1ff87cfc40a011a63db574edad7e6
    - GIT_PREVIOUS_COMMIT:5d02b46fdb653f789e799ff6ad304baccc32cbf9
    - GIT_PREVIOUS_SUCCESSFUL_COMMIT:5d02b46fdb653f789e799ff6ad304baccc32cbf9
    - GIT_URL:https://github.com/vegaprotocol/data-node.git
*/
def scmVars = null
def version = 'UNKNOWN'
def versionHash = 'UNKNOWN'
def commitHash = 'UNKNOWN'


pipeline {
    agent any
    options {
        skipDefaultCheckout true
        timestamps()
        timeout(time: 45, unit: 'MINUTES')
    }
    parameters {
        string( name: 'VEGA_CORE_BRANCH', defaultValue: '',
                description: '''Git branch, tag or hash of the vegaprotocol/vega repository.
                    e.g. "develop", "v0.44.0 or commit hash. Default empty: use latests published version.''')
        string( name: 'VEGAWALLET_BRANCH', defaultValue: '',
                description: '''Git branch, tag or hash of the vegaprotocol/vegawallet repository.
                    e.g. "develop", "v0.9.0" or commit hash. Default empty: use latest published version.''')
        string( name: 'ETHEREUM_EVENT_FORWARDER_BRANCH', defaultValue: '',
                description: '''Git branch, tag or hash of the vegaprotocol/ethereum-event-forwarder repository.
                    e.g. "main", "v0.44.0" or commit hash. Default empty: use latest published version.''')
        string( name: 'DEVOPS_INFRA_BRANCH', defaultValue: 'master',
                description: 'Git branch, tag or hash of the vegaprotocol/devops-infra repository')
        string( name: 'VEGATOOLS_BRANCH', defaultValue: 'develop',
                description: 'Git branch, tag or hash of the vegaprotocol/vegatools repository')
        string( name: 'SYSTEM_TESTS_BRANCH', defaultValue: 'develop',
                description: 'Git branch, tag or hash of the vegaprotocol/system-tests repository')
        string( name: 'PROTOS_BRANCH', defaultValue: 'develop',
                description: 'Git branch, tag or hash of the vegaprotocol/protos repository')
    }
    environment {
        GO111MODULE = 'on'
        CGO_ENABLED  = '0'
        DOCKER_IMAGE_TAG_LOCAL = "j-${ env.JOB_BASE_NAME.replaceAll('[^A-Za-z0-9\\._]','-') }-${BUILD_NUMBER}-${EXECUTOR_NUMBER}"
        DOCKER_IMAGE_NAME_LOCAL = "docker.pkg.github.com/vegaprotocol/data-node/data-node:${DOCKER_IMAGE_TAG_LOCAL}"
    }

    stages {
        stage('Config') {
            steps {
                cleanWs()
                sh 'printenv'
                echo "${params}"
            }
        }
        stage('Git Clone') {
            options { retry(3) }
            steps {
                script {
                    scmVars = checkout(scm)
                    versionHash = sh (returnStdout: true, script: "echo \"${scmVars.GIT_COMMIT}\"|cut -b1-8").trim()
                    version = sh (returnStdout: true, script: "git describe --tags 2>/dev/null || echo ${versionHash}").trim()
                    commitHash = getCommitHash()
                }
                echo "scmVars=${scmVars}"
                echo "commitHash=${commitHash}"
            }
        }

        stage('Dependencies') {
            options { retry(3) }
            steps {
                sh 'go mod download -x'
            }
        }

        stage('Run tests') {
            parallel {
                stage('unit tests') {
                    options { retry(3) }
                    steps {
                        sh 'go test -v $(go list ./... | grep -v api_test) 2>&1 | tee unit-test-results.txt && cat unit-test-results.txt | go-junit-report > vega-unit-test-report.xml'
                        junit checksName: 'Unit Tests', testResults: 'vega-unit-test-report.xml'
                    }
                }
                stage('unit tests with race') {
                    environment {
                        CGO_ENABLED = '1'
                    }
                    options { retry(3) }
                    steps {
                        sh 'go test -v -race $(go list ./... | grep -v api_test) 2>&1 | tee unit-test-race-results.txt && cat unit-test-race-results.txt | go-junit-report > vega-unit-test-race-report.xml'
                        junit checksName: 'Unit Tests with Race', testResults: 'vega-unit-test-race-report.xml'
                    }
                }
            }
        }

        stage('Publish') {
            parallel {

                stage('docker image') {
                    when {
                        anyOf {
                            buildingTag()
                            branch 'develop'
                            // changeRequest() // uncomment only for testing
                        }
                    }
                    environment {
                        DOCKER_IMAGE_TAG_VERSIONED = "${ env.TAG_NAME ? env.TAG_NAME : env.BRANCH_NAME }"
                        DOCKER_IMAGE_NAME_VERSIONED = "docker.pkg.github.com/vegaprotocol/data-node/data-node:${DOCKER_IMAGE_TAG_VERSIONED}"
                        DOCKER_IMAGE_TAG_ALIAS = "${ env.TAG_NAME ? 'latest' : 'edge' }"
                        DOCKER_IMAGE_NAME_ALIAS = "docker.pkg.github.com/vegaprotocol/data-node/data-node:${DOCKER_IMAGE_TAG_ALIAS}"
                    }
                    options { retry(3) }
                    steps {
                        sh label: 'Tag new images', script: '''#!/bin/bash -e
                            docker image tag "${DOCKER_IMAGE_NAME_LOCAL}" "${DOCKER_IMAGE_NAME_VERSIONED}"
                            docker image tag "${DOCKER_IMAGE_NAME_LOCAL}" "${DOCKER_IMAGE_NAME_ALIAS}"
                        '''

                        withDockerRegistry([credentialsId: 'github-vega-ci-bot-artifacts', url: "https://docker.pkg.github.com"]) {
                            sh label: 'Push docker images', script: '''
                                docker push "${DOCKER_IMAGE_NAME_VERSIONED}"
                                docker push "${DOCKER_IMAGE_NAME_ALIAS}"
                            '''
                        }
                        slackSend(
                            channel: "#tradingcore-notify",
                            color: "good",
                            message: ":docker: Data-Node » Published new docker image `${DOCKER_IMAGE_NAME_VERSIONED}` aka `${DOCKER_IMAGE_NAME_ALIAS}`",
                        )
                    }
                }

                stage('release to GitHub') {
                    when {
                        buildingTag()
                    }
                    environment {
                        RELEASE_URL = "https://github.com/vegaprotocol/data-node/releases/tag/${TAG_NAME}"
                    }
                    options { retry(3) }
                    steps {
                        withCredentials([usernamePassword(credentialsId: 'github-vega-ci-bot-artifacts', passwordVariable: 'TOKEN', usernameVariable:'USER')]) {
                            // Workaround for user input:
                            //  - global configuration: 'gh config set prompt disabled'
                            sh label: 'Log in to a Gihub with CI', script: '''
                                echo ${TOKEN} | gh auth login --with-token -h github.com
                            '''
                        }
                        sh label: 'Upload artifacts', script: '''#!/bin/bash -e
                            [[ $TAG_NAME =~ '-pre' ]] && prerelease='--prerelease' || prerelease=''
                            gh release create $TAG_NAME $prerelease ./cmd/data-node/data-node-*
                        '''
                        slackSend(
                            channel: "#tradingcore-notify",
                            color: "good",
                            message: ":rocket: Data-Node » Published new version to GitHub <${RELEASE_URL}|${TAG_NAME}>",
                        )
                    }
                    post {
                        always  {
                            retry(3) {
                                script {
                                    sh label: 'Log out from Github', script: '''
                                        gh auth logout -h github.com
                                    '''
                                }
                            }
                        }
                    }
                }

                stage('[TODO] deploy to Devnet') {
                    when {
                        branch 'develop'
                    }
                    steps {
                        echo 'Deploying to Devnet....'
                        echo 'Run basic tests on Devnet network ...'
                    }
                }
            }
        }

    }
    post {
        cleanup {
            retry(3) {
                sh label: 'Clean docker images', script: '''
                    docker rmi "${DOCKER_IMAGE_NAME_LOCAL}"
                '''
            }
        }
    }
}
