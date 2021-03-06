image = gcr.io/incentro-oss/kube-utilize-operator
version = alpha1-7

cc:
	kubectl config current-context

b:
	operator-sdk generate k8s
	operator-sdk build $(image):$(version)

p:
	docker push $(image):$(version)

# clean operator
co:
	kubectl delete -f deploy/crds/utilize.incentro.com_utilizesets_crd.yaml
	kubectl delete -f deploy/operator.yaml
	kubectl delete -f deploy/role_binding.yaml
	kubectl delete -f deploy/role.yaml
	kubectl delete -f deploy/service_account.yaml

# deploy operator
do:
	kubectl apply -f deploy/role_binding.yaml
	kubectl apply -f deploy/role.yaml
	kubectl apply -f deploy/service_account.yaml
	kubectl apply -f deploy/cluster_role.yaml
	kubectl apply -f deploy/cluster_role_binding.yaml
	kubectl apply -f deploy/crds/utilize.incentro.com_utilizesets_crd.yaml
	sed -i 's|REPLACE_IMAGE|$(image):$(version)|g' deploy/operator.yaml
	kubectl apply -f deploy/operator.yaml
	sed -i 's|$(image):$(version)|REPLACE_IMAGE|g' deploy/operator.yaml

# test CR
tcr:
	kubectl apply -f deploy/crds/utilize.incentro.com_v1alpha1_utilizeset_cr.yaml

# clean CR
ccr:
	kubectl delete -f deploy/crds/utilize.incentro.com_v1alpha1_utilizeset_cr.yaml
