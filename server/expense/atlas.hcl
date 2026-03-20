env "local" {
  src = "ent://internal/infra/ent/schema"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
}
