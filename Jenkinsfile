pipeline {

  agent {
    label "golangci-lint && golang"
  }


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
            sh 'golangci-lint run'
          }
        }
      }
    }
  }
}
