package radikocast

import (
	"os"
	"testing"
)

func TestAWSRegionUsesEnvironmentFallback(t *testing.T) {
	old := os.Getenv("AWS_REGION")
	defer func() { _ = os.Setenv("AWS_REGION", old) }()
	if err := os.Setenv("AWS_REGION", "ap-northeast-1"); err != nil {
		t.Fatalf("set AWS_REGION: %v", err)
	}
	if got := awsRegion(); got != "ap-northeast-1" {
		t.Fatalf("awsRegion() = %q, want ap-northeast-1", got)
	}

	if err := os.Unsetenv("AWS_REGION"); err != nil {
		t.Fatalf("unset AWS_REGION: %v", err)
	}
	if err := os.Setenv("AWS_DEFAULT_REGION", "us-east-1"); err != nil {
		t.Fatalf("set AWS_DEFAULT_REGION: %v", err)
	}
	if got := awsRegion(); got != "us-east-1" {
		t.Fatalf("awsRegion() = %q, want us-east-1", got)
	}
}
