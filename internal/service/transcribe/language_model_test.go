package transcribe_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftranscribe "github.com/hashicorp/terraform-provider-aws/internal/service/transcribe"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTranscribeLanguageModel_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var languageModel types.LanguageModel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_language_model.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.TranscribeEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.TranscribeEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLanguageModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLanguageModelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLanguageModelExists(resourceName, &languageModel),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "transcribe", regexp.MustCompile(`languagemodel:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccTranscribeLanguageModel_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var languageModel types.LanguageModel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transcribe_languagemodel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.TranscribeEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.TranscribeEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLanguageModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLanguageModelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLanguageModelExists(resourceName, &languageModel),
					acctest.CheckResourceDisappears(acctest.Provider, tftranscribe.ResourceLanguageModel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLanguageModelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transcribe_language_model" {
			continue
		}

		_, err := tftranscribe.FindLanguageModelByName(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLanguageModelExists(name string, languageModel *types.LanguageModel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transcribe LanguageModel is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn
		resp, err := tftranscribe.FindLanguageModelByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("Error describing Transcribe LanguageModel: %s", err.Error())
		}

		*languageModel = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TranscribeConn

	input := &transcribe.ListLanguageModelsInput{}

	_, err := conn.ListLanguageModels(context.Background(), input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccLanguageModelBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}
func testAccLanguageModelConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_transcribe_languagemodel" "test" {
  languagemodel_name             = %[1]q
  engine_type             = "ActiveTranscribe"
  engine_version          = %[2]q
  host_instance_type      = "transcribe.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}
