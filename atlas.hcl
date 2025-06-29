variable "postgres_host" {
  type = string
  default = getenv("POSTGRES_HOST")
}

variable "postgres_port" {
  type = string
  default = getenv("POSTGRES_PORT")
}

variable "postgres_user" {
  type = string
  default = getenv("POSTGRES_USER")
}

variable "postgres_password" {
  type = string
  default = getenv("POSTGRES_PASSWORD")
}

variable "postgres_db" {
  type = string
  default = getenv("POSTGRES_DB")
}

variable "postgres_sslmode" {
  type = string
  default = getenv("POSTGRES_SSLMODE")
}

env "local" {
  url = "postgres://${var.postgres_user}:${var.postgres_password}@${var.postgres_host}:${var.postgres_port}/${var.postgres_db}?sslmode=${var.postgres_sslmode}"
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
} 

env "prod" {
  url = "postgres://${var.postgres_user}:${var.postgres_password}@${var.postgres_host}:${var.postgres_port}/${var.postgres_db}?sslmode=${var.postgres_sslmode}"
}