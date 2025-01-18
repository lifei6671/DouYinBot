package storage

type Options struct {
	BucketName      string `json:"bucket_name"`
	AccountID       string `json:"account_id"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Endpoint        string `json:"endpoint"`
	Domain          string `json:"domain"`
}

type OptionsFunc func(*Options) error

// WithBucketName 设置Bucket名称
func WithBucketName(bucketName string) OptionsFunc {
	return func(opt *Options) error {
		opt.BucketName = bucketName
		return nil
	}
}

// WithAccountID 设置账号
func WithAccountID(accountID string) OptionsFunc {
	return func(opt *Options) error {
		opt.AccountID = accountID
		return nil
	}
}
func WithAccessKeyID(accessKeyID string) OptionsFunc {
	return func(opt *Options) error {
		opt.AccessKeyID = accessKeyID
		return nil
	}
}
func WithAccessKeySecret(accessKeySecret string) OptionsFunc {
	return func(opt *Options) error {
		opt.AccessKeySecret = accessKeySecret
		return nil
	}
}
func WithDomain(domain string) OptionsFunc {
	return func(opt *Options) error {
		opt.Domain = domain
		return nil
	}
}

func WithEndpoint(endpoint string) OptionsFunc {
	return func(opt *Options) error {
		opt.Endpoint = endpoint
		return nil
	}
}
