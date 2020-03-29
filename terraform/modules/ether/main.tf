data "aws_region" "ether" {}

resource "aws_cloudwatch_log_group" "ether" {
  name              = "${var.name}_ether"
  retention_in_days = 1
}

resource "aws_ecs_task_definition" "ether" {
  family       = "${var.name}_ether"
  network_mode = "bridge"

  container_definitions = <<EOF
[
  {
    "name": "${var.name}_ether",
    "image": "343660461351.dkr.ecr.us-east-2.amazonaws.com/ether:${var.container_tag}",
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
            "awslogs-group": "${aws_cloudwatch_log_group.ether.name}",
            "awslogs-region": "${data.aws_region.ether.name}",
            "awslogs-stream-prefix": "${var.name}"
        }
    },
    "cpu": 10,
    "memory": 128,
    "essential": true,
    "environment": [
        {
            "name": "ETHER_DB_LOCATION",
            "value": "${var.db_location}"
        },
        {
            "name": "ETHER_DB_USERNAME",
            "value": "${var.db_username}"
        },
        {
            "name": "ETHER_DB_PASSWORD",
            "value": "${var.db_password}"
        },
        {
            "name": "ETHER_CONTENT_DIR",
            "value": "/tmp"
        },
        {
            "name": "ETHER_KAFKA_SERVER",
            "value": "${var.kafka_server}"
        },
        {
            "name": "ETHER_KAFKA_TOPIC",
            "value": "${var.kafka_topic}"
        }
    ],
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": ${var.port},
        "protocol": "tcp"
      }
    ],
    "mountPoints": [
      {
        "sourceVolume": "efsVolume",
        "containerPath": "/tmp"
      }
    ]
  }
]
EOF

  volume {
    name = "efsVolume"
    efs_volume_configuration {
      file_system_id = var.efs_id
      root_directory = "/"
    }
  }
}

resource "aws_elb" "ether" {
  name            = "${var.name}-ether"
  subnets         = var.subnets
  security_groups = var.security_groups
  internal        = var.internal

  listener {
    instance_port     = var.port
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_ecs_service" "ether" {
  name            = "${var.name}_ether"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.ether.arn

  load_balancer {
    elb_name       = aws_elb.ether.name
    container_name = "${var.name}_ether"
    container_port = 80
  }

  desired_count = 1
}
