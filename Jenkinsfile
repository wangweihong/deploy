node("master") {
	sh 'rm -rf *'
		//代码检出
		stage('Check out') {
			git credentialsId: 'ufleet-deploy-git', url: 'http://192.168.19.250/ufleet/ufleet-deploy.git'
		} 
	// 镜像中代码构建
	stage('Build'){
		def confFilePath = 'conf/app.conf'
			def config = readFile confFilePath
			config = config.replaceFirst(/runmode = dev/, "runmode = prod")  
			writeFile file: confFilePath, text: config

			docker.image('192.168.18.250:5002/ufleet-build/golang:1.7').inside {
				sh './script/build.sh'
			}
	} 
	// 编译镜像并push到仓库
	stage('Image Build And Push'){
	def imageTag = "v1.7.0.${BUILD_NUMBER}"
	def dockerfile = readFile 'Dockerfile'
	dockerfile = dockerfile.replaceFirst(/# ENV MODULE_VERSION #MODULE_VERSION#/, "ENV MODULE_VERSION ${imageTag}")
	writeFile file: 'Dockerfile', text: dockerfile

		docker.withRegistry('http://192.168.18.250:5002', '18.250-registry-admin') {
			//			docker.build('192.168.18.250:5002/ufleet/ufleet-deploy:v0.1.${BUILD_NUMBER}').push()
			docker.build('192.168.18.250:5002/ufleet/ufleet-deploy:'+imageTag).push()
	}
	}
}
