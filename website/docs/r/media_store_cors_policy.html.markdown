---
layout: "aws"
page_title: "AWS: aws_media_store_cors_policy"
sidebar_current: "docs-aws-resource-media-store-cors-policy"
description: |-
  Provides a MediaStore Cors Policy.
---

# aws_media_store_cors_policy

Provides a MediaStore Cors Policy.

## Example Usage

```hcl
resource "aws_media_store_container" "example" {
  name = "example"
}

resource "aws_media_store_cors_policy" "example" {
  container_name = "${aws_media_store_container.example.name}"
  cors_policy {
    allowed_headers = ["*"]
    allowed_origins = ["*"]
    allowed_methods = ["GET"]
    expose_headers = ["*"]
    max_age_seconds = 3000
  }
}
```

## Argument Reference

The following arguments are supported:

* `container_name` - (Required) The name of the container.
* `cors_policy` - (Required) The cors policy.

### `cors_policy`

* `allowed_headers` - (Optional) Specifies which headers are allowed in a preflight OPTIONS request through the Access-Control-Request-Headers header.
* `allowed_origins` - (Optional) One or more response headers that you want users to be able to access from their applications.
* `allowed_methods` - (Optional) Identifies an HTTP method that the origin that is specified in the rule is allowed to execute.
* `expose_headers` - (Optional) One or more headers in the response that you want users to be able to access from their applications.
* `max_age_seconds` - (Optional) The time in seconds that your browser caches the preflight response for the specified resource.

## Import

MediaStore Cors Policy can be imported using the MediaStore Container Name, e.g.

```
$ terraform import aws_media_store_cors_policy.example example
```
