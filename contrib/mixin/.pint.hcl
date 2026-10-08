// CI validates generated rules without contacting a Prometheus server.
rule {
  match {
    kind = "alerting"
  }

  label "severity" {
    required = true
    value    = ".+"
    severity = "bug"
  }

  annotation "summary" {
    required = true
    value    = ".+"
    severity = "bug"
  }

  annotation "description" {
    required = true
    value    = ".+"
    severity = "bug"
  }
}
