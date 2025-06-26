env "local" {
  url = "sqlite://data/radio.db"
  dev = "sqlite://data/radio.db"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
} 