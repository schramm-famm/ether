variable "name" {
  type        = string
  description = "Name used to identify resources"
}

variable "access_key" {
  type        = string
  description = "AWS access key ID"
}

variable "secret_key" {
  type        = string
  description = "AWS secret access key"
}

variable "region" {
  type        = string
  description = "AWS region to deploy where resources will be deployed"
  default     = "us-east-2"
}

variable "rds_username" {
  type        = string
  description = "Username for the master RDS user"
}

variable "rds_password" {
  type        = string
  description = "Password for the master RDS user"
}

variable "ether_container_tag" {
  type        = string
  description = "Tag of the Docker container to be used in the ether container definition"
  default     = "latest"
}

variable "karen_container_tag" {
  type        = string
  description = "Tag of the Docker container to be used in the karen container definition"
  default     = "latest"
}
