package cmd

func askForMasterSecret() string {
	sqlog.Print(styleNotice.Render("Squonker sync operations require a master password, used to more safely store other credentials."))
	sqlog.Print(styleNotice.Render("Please enter & remember your master password carefully, it will not be echoed to the screen, nor stored anywhere."))
	secretPhrase, err := readPwdFromTerminal("sync-master-pass")
	sqlog.Print("")
	if err != nil {
		sqFatal(err)
	}
	if len(secretPhrase) < 1 {
		sqFatal("sync-master-pass must be at least 1 character")
	}
	return secretPhrase
}
