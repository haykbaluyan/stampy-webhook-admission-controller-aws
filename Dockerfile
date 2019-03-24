FROM dva-registry.internal.salesforce.com/dva/sfdc_centos7:38

ADD bin/stampy-webhook-admission-controller-aws /stampy-webhook-admission-controller-aws
ENTRYPOINT ["/stampy-webhook-admission-controller-aws"]