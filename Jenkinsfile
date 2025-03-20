pipeline{
    environment{
        DOCKER_CREDENTIALS = credentials("docker-credentials")
        registry = "csye712504/db-backup-operator"
        GIT_CREDENTIALS = credentials("git-credentials")
        // Initialize default version
        CURRENT_VERSION = "0.0.0"
    }
    agent any
    stages{
        stage("Get Current Version") {
            steps {
                script {
                    GIT_REPO_URL = sh(
                        script: 'git config --get remote.origin.url',
                        returnStdout: true
                    ).trim()
                    
                    // Extract owner and repo name from URL
                    // Handle both HTTPS and SSH URLs
                    if (GIT_REPO_URL.startsWith('https://')) {
                        def matcher = GIT_REPO_URL =~ /https:\/\/github.com\/([^\/]+)\/([^\/\.]+)(\.git)?$/
                        if (matcher.matches()) {
                            REPO_OWNER = matcher[0][1]
                            REPO_NAME = matcher[0][2]
                        }
                    } else {
                        // Handle SSH URL format: git@github.com:owner/repo.git
                        def matcher = GIT_REPO_URL =~ /git@github.com:([^\/]+)\/([^\/\.]+)(\.git)?$/
                        if (matcher.matches()) {
                            REPO_OWNER = matcher[0][1]
                            REPO_NAME = matcher[0][2]
                        }
                    }
                    
                    println "Full path: ${REPO_OWNER}/${REPO_NAME}"

                    withCredentials([string(credentialsId: 'shalom-PAT', variable: 'GITHUB_TOKEN')]) {
                        try {
                            // Run the shell command to delete all local git tags
                            sh 'git tag -l | xargs git tag -d'
                            echo "Successfully deleted all local tags."
                        } catch (Exception e) {
                            echo "Error deleting local tags: ${e.message}"
                        }

                        try {
                            // sh ''' git remote set-url origin https://$GITHUB_TOKEN@github.com/$REPO_OWNER/$REPO_NAME.git '''
                            sh 'git remote set-url origin https://$GITHUB_TOKEN@github.com/cyse7125-sp25-team04/db-backup-operator.git'
                            // Fetch all remotes and tags
                            sh 'git fetch --all --tags'
                            echo "Successfully fetched all tags from all remotes."
                        } catch (Exception e) {
                            echo "Error fetching tags: ${e.message}"
                        }
                        // Get the latest tag from git and store in latestTag variable
                        def latestTag = sh(
                            script: "git tag -l --sort=-version:refname | head -n 1",
                            returnStdout: true
                        ).trim()

                        // if no tag is present in the repo
                        if (!latestTag.matches(/^v\d+\.\d+\.\d+$/)) {
                            latestTag = "v0.0.0"
                        }
                        CURRENT_VERSION = latestTag

                        // Print the latest tag
                        echo "Latest Tag: ${CURRENT_VERSION}"
                    }
                }
            }
        }
        
        stage("Determine Version Bump") {
            steps {
                script {
                    withCredentials([string(credentialsId: 'shalom-PAT', variable: 'GITHUB_TOKEN')]) {
                        // Get the latest commit message
                        def commitMessage = sh(
                            script: "git log -1 --pretty=%B",
                            returnStdout: true
                        ).trim().toLowerCase()
                        
                        def (major, minor, patch) = CURRENT_VERSION.tokenize('.')
                        def newVersion = ""
                        
                        // Convert commit message to lowercase for consistent comparison
                        commitMessage = commitMessage.toLowerCase()

                        // Parse commit message and increment version accordingly
                        if (commitMessage.startsWith('breaking change:') || commitMessage.contains('!:')) {
                            newVersion = "${(major.toInteger() + 1)}.0.0"
                        } else if (commitMessage.startsWith('feat:')) {
                            newVersion = "${major}.${(minor.toInteger() + 1)}.0"
                        } else if (commitMessage.startsWith('fix:') || 
                                commitMessage.startsWith('build:') || 
                                commitMessage.startsWith('ci:') || 
                                commitMessage.startsWith('docs:') || 
                                commitMessage.startsWith('style:') || 
                                commitMessage.startsWith('perf:')) {
                            newVersion = "${major}.${minor}.${(patch.toInteger() + 1)}"
                        } else {
                            newVersion = "${major}.${minor}.${(patch.toInteger() + 1)}"
                        }
                        
                        // Store the new version
                        env.NEW_VERSION = newVersion
                        
                        // Create and push git tag
                        sh """
                            git tag -a ${NEW_VERSION} -m "Release version ${NEW_VERSION}"
                            git push https://$GITHUB_TOKEN@github.com/${REPO_OWNER}/${REPO_NAME}.git ${NEW_VERSION}
                        """
                    }
                }
            }
        }

        stage("Image Building and Pushing") {
            steps {
                script {
                    // Login to Docker registry
                    sh 'echo ${DOCKER_CREDENTIALS_PSW} | docker login -u ${DOCKER_CREDENTIALS_USR} --password-stdin'
                    // Create and use a builder instance
                    sh 'docker buildx create --use --name imagebuilder'
                    
                    // 1) Build and push the API image from the api folder, then cd back out
                    sh """
                        go mod tidy
                        make generate
                        docker buildx build --platform linux/amd64,linux/arm64 \\
                            -t ${registry}:latest \\
                            -t ${api_registry}:${NEW_VERSION} \\
                            --push .
                    """

                    // Remove the builder instance
                    sh 'docker buildx rm imagebuilder'
                }
            }
        }

        stage('Clean Workspace') {
            steps {
                script {
                    // Clean workspace at the start to ensure no leftover files
                    deleteDir()
                }
            }
        }

    }

    post {
        failure {
            echo "Build failed. Cleaning up builder instance."
            sh 'docker buildx rm imagebuilder || true'
        }
        always {
            echo "Pipeline execution completed."
        }
    }
}
