FROM dva-registry.internal.salesforce.com/dva/sfdc_centos7:38

ADD bin/stampy-admission-webhook /stampy-admission-webhook
ENTRYPOINT ["/stampy-admission-webhook"]