package main

const VERSION = "0.1"

// clientVersion tries to print version of client
func clientVersion() (string, error) {
	// TODO: this should be more sophisticated (version from git tag)
	return VERSION, nil
}

// serverVersion tries to get version of server and version of rules
func serverVersion() (*string, *string, error) {
	rhsmStatus, err := rhsmClient.GetServerStatus()
	if err != nil {
		return nil, nil, err
	}

	serverVersionRelease := rhsmStatus.Version + rhsmStatus.Release
	return &serverVersionRelease, &rhsmStatus.RulesVersion, nil

}
