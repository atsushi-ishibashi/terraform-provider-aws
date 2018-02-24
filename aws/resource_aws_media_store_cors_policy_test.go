package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsMediaStoreCorsPolicy_basic(t *testing.T) {
	rname := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreCorsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreCorsPolicyConfig_basic(rname),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreCorsPolicyExists("aws_media_store_cors_policy.test"),
					resource.TestCheckResourceAttrSet("aws_media_store_cors_policy.test", "container_name"),
					resource.TestCheckResourceAttr("aws_media_store_cors_policy.test", "cors_policy.#", "1"),
				),
			},
			{
				Config: testAccMediaStoreCorsPolicyConfig_update(rname),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreCorsPolicyExists("aws_media_store_cors_policy.test"),
					resource.TestCheckResourceAttrSet("aws_media_store_cors_policy.test", "container_name"),
					resource.TestCheckResourceAttr("aws_media_store_cors_policy.test", "cors_policy.#", "2"),
				),
			},
		},
	})
}

func TestAccAwsMediaStoreCorsPolicy_import(t *testing.T) {
	resourceName := "aws_media_store_cors_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreCorsPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccMediaStoreCorsPolicyConfig_basic(acctest.RandString(5)),
			},
			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMediaStoreCorsPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_cors_policy" {
			continue
		}

		input := &mediastore.GetCorsPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCorsPolicy(input)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, mediastore.ErrCodeCorsPolicyNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MediaStore Container Policy to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAwsMediaStoreCorsPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

		input := &mediastore.GetCorsPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCorsPolicy(input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccMediaStoreCorsPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}

resource "aws_media_store_cors_policy" "test" {
  container_name = "${aws_media_store_container.test.name}"
  cors_policy {
    allowed_headers = ["*"]
    allowed_origins = ["*"]
    allowed_methods = ["HEAD"]
    expose_headers = ["*"]
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccMediaStoreCorsPolicyConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}

resource "aws_media_store_cors_policy" "test" {
  container_name = "${aws_media_store_container.test.name}"
  cors_policy {
    allowed_headers = ["*"]
    allowed_origins = ["*"]
    allowed_methods = ["GET"]
    expose_headers = ["*"]
    max_age_seconds = 3000
  }
  cors_policy {
    allowed_headers = ["*"]
    allowed_origins = ["*"]
    allowed_methods = ["GET"]
    expose_headers = ["*"]
    max_age_seconds = 1000
  }
}
`, rName)
}
