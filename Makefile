
.PHONY:

#使用code-generator生成自定义资源操作的相关代码
gen:
	cd ./k8s/crd && ./hack/update-codegen.sh

#生成接口文档
swag: .PHONY
	./swag init

#拉取最新的前端代码
front: .PHONY
	rm -rf cloudApp-front
	git clone https://gitee.com/coolsun972/cloudApp-front
	rm -rf ./static/img && rm -rf ./static/css && rm -rf ./static/js
	rm -rf ./view/index.html
	mv  cloudApp-front/cloudApp/dist/static/img ./static
	mv  cloudApp-front/cloudApp/dist/static/css ./static
	mv  cloudApp-front/cloudApp/dist/static/js ./static
	mv  cloudApp-front/cloudApp/dist/index.html view
	rm -rf cloudApp-front


build: .PHONY front
	docker build -t coolsun972/cloud-app:$(version) .
	docker push coolsun972/cloud-app:$(version)

test: .PHONY
	@echo $(version)
