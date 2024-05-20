package eureka

type EurekaClient struct {

}

type EurekaClient struct {

}

//func NewEurekaDiscoveryClient(url, appId string) DiscoveryClient {
//	client := eureka.NewClient(&eureka.Config{
//		DefaultZone:                  url, //"http://localhost:8761/eureka/",
//		App:                          appId,
//		Port:                         10000,
//		RenewalIntervalInSecs:        10,
//		RegistryFetchIntervalSeconds: 15,
//		DurationInSecs:               30,
//		Metadata: map[string]interface{}{
//			"VERSION":              "0.1.0",
//			"NODE_GROUP_ID":        0,
//			"PRODUCT_CODE":         "DEFAULT",
//			"PRODUCT_VERSION_CODE": "DEFAULT",
//			"PRODUCT_ENV_CODE":     "DEFAULT",
//			"SERVICE_VERSION_CODE": "DEFAULT",
//		},
//	})
//	//, func(instance *eureka.Instance) {
//	//	// custom instance
//	//	instance.InstanceID = "go-example"
//	//}
//	return &EurekaDiscoveryClient{
//		client: client,
//	}
//}
