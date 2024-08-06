package thing2

func modelsInstall() {
	for _, maker := range Makers {
		proto := maker("proto", "proto")
		proto.InstallModel()
	}
}
