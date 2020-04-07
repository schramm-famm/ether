variable "name" {
  type        = string
  description = "Name used to identify resources"
}

variable "container_tag" {
  type        = string
  description = "Tag of the ether container in the registry to be used"
  default     = "latest"
}

variable "port" {
  type        = number
  description = "The port that ether's container port will map to on the host"
  default     = 80
}

variable "container_count" {
  type        = number
  description = "The number of containers to deploy in the ether service"
  default     = 1
}

variable "cluster_id" {
  type        = string
  description = "ID of the ECS cluster that the ether service will run in"
}

variable "security_groups" {
  type        = list(string)
  description = "VPC security groups for the ether service load balancer"
}

variable "subnets" {
  type        = list(string)
  description = "VPC subnets for the ether service load balancer"
}

variable "internal" {
  type        = bool
  description = "Toggle whether the load balancer will be internal"
}

variable "db_location" {
  type        = string
  description = "Location (host) of the MariaDB server"
}

variable "db_username" {
  type        = string
  description = "Username for accessing the MariaDB server"
}

variable "db_password" {
  type        = string
  description = "Password for accessing the MariaDB server"
}

variable "kafka_server" {
  type        = string
  description = "Server where Kafka is running"
}

variable "kafka_topic" {
  type        = string
  description = "Kafka topic to read from"
}

variable "karen_endpoint" {
  type        = string
  description = "Endpoint for accessing the karen service"
}

variable "efs_id" {
  type        = string
  description = "ID of the EFS file system mounted on the container instances"
}
