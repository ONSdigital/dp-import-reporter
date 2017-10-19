job "dp-import-reporter" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  // Make sure that this API is only ran on the publishing nodes
  constraint {
    attribute = "${node.class}"
    value     = "publishing"
  }

  group "publishing" {
    count = "{{PUBLISHING_TASK_COUNT}}"

    task "dp-import-reporter" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{BUILD_BUCKET}}/dp-import-reporter/{{REVISION}}.tar.gz"
      }

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-import-reporter/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args    = [
          "${NOMAD_TASK_DIR}/dp-import-reporter"
        ]
      }

      service {
        name = "dp-import-reporter"
        tags = ["publishing"]
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-import-reporter"]
      }
    }
  }
}