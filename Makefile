.PHONY: kube
kube:
	@kubectl apply -f tools/$(folder)/$(file)