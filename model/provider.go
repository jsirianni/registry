package model

// ProviderVersions represents a provider's version json file
type ProviderVersions struct {
	Versions []ProviderVersion `json:"versions"`
}

// ProviderVersion represents a provider version
type ProviderVersion struct {
	Version   string   `json:"version"`
	Protocols []string `json:"protocols"`
	Platforms []struct {
		Os                  string `json:"os"`
		Arch                string `json:"arch"`
		Filename            string `json:"filename"`
		DownloadURL         string `json:"download_url"`
		ShasumsURL          string `json:"shasums_url"`
		ShasumsSignatureURL string `json:"shasums_signature_url"`
		Shasum              string `json:"shasum"`
		SigningKeys         struct {
			GpgPublicKeys []struct {
				KeyID          string `json:"key_id"`
				ASCIIArmor     string `json:"ascii_armor"`
				TrustSignature string `json:"trust_signature"`
				Source         string `json:"source"`
				SourceURL      string `json:"source_url"`
			} `json:"gpg_public_keys"`
		} `json:"signing_keys"`
	} `json:"platforms"`
}

// DownloadResponse represents the download endpoints response body
type DownloadResponse struct {
	Protocols           []string `json:"protocols"`
	Os                  string   `json:"os"`
	Arch                string   `json:"arch"`
	Filename            string   `json:"filename"`
	DownloadURL         string   `json:"download_url"`
	ShasumsURL          string   `json:"shasums_url"`
	ShasumsSignatureURL string   `json:"shasums_signature_url"`
	Shasum              string   `json:"shasum"`
	SigningKeys         struct {
		GpgPublicKeys []struct {
			KeyID          string `json:"key_id"`
			ASCIIArmor     string `json:"ascii_armor"`
			TrustSignature string `json:"trust_signature"`
			Source         string `json:"source"`
			SourceURL      string `json:"source_url"`
		} `json:"gpg_public_keys"`
	} `json:"signing_keys"`
}

// Version represents a provider version's supported
// protocols and platforms
type Version struct {
	Version   string   `json:"version"`
	Protocols []string `json:"protocols"`
	Platforms []struct {
		Os   string `json:"os"`
		Arch string `json:"arch"`
	} `json:"platforms"`
}
