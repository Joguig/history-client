pipeline {
  agent any
  options {
    timestamps()
    ansiColor('xterm')
  }

  environment {
    PKG = 'code.justin.tv/foundation/history.v2'
  }

  stages {
    stage('build') {
      agent { docker { image 'docker.internal.justin.tv/devtools/xenial/go1.11.4' } }
      steps {
        sh 'mkdir -p $(dirname $GOPATH/src/$PKG)'
        sh 'cp -r $(pwd) $GOPATH/src/$PKG'
        sh 'go get github.com/twitchtv/retool'
        sh 'cd $GOPATH/src/$PKG && dep ensure'
        sh 'cd $GOPATH/src/$PKG && retool do gometalinter ./...'
        sh 'cd $GOPATH/src/$PKG && go test -short ./...'
      }
    }
  }
}
