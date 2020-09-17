package global

var BuildEnvironment = new(buildEnvironment)

type buildEnvironment struct {

	/** Per-project directory for final artifacts */
	OutDir string

	/** Per-project working directory */
	WorkDir string

	/** Global/cross-project cache directory */
	CacheDir string
}
