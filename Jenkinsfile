pipeline {

  agent any

  tools { go 'go-1.23.3'}

  stages {
    stage('CI') {
      parallel {
        stage('test') {
          steps {
            sh 'go test -cover -v ./...'
          }
        }
        stage('lint') {
          steps {
            sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.2.2'
            sh './bin/golangci-lint run'
          }
        }
      }
    }
  }
}
