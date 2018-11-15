package conf

// Encryption defines how a projection mapping wants to encrypt its keys
type Encryption struct {
	Module                string `yaml:"module",json:"module"`
	IncludeDecryptionKeys bool   `yaml:"include_decryption_keys",json:"include_decryption_keys"`
	// PluginPath is the path to the .so on the filesystem, if this Module is loaded from a shared object, and Module: "plugin"
	PluginPath string `yaml:"plugin-path",json:"plugin-path"`
	// Options are arbitrary flags available to underlying implementations
	Params map[string]string `yaml:"params",json:"params"`
	// CredsKeysFilePath tends to not be specified in a projection mapping; this is merged from the CLI flags
	CredsKeysFilePath string `yaml:"creds_keys_file",json:"creds_keys_file"`
	// KeysDecrypterFilePath tends to not be specified in a projection mapping; this is merged from the CLI flags
	KeysDecrypterFilePath string `yaml:"keys_decrypter_file",json:"keys_decrypter_file"`
}
