resource "autodns_record" "example_A" {
  zone_id = "foobar.test@bar.ns.net"

  name   = "foo"
  ttl    = 60
  type   = "A"
  values = ["1.1.1.1"]
}

resource "autodns_record" "example_TXT" {
  zone_id = "foobar.test@bar.ns.net"

  name   = "foo"
  ttl    = 60
  type   = "TXT"
  values = ["foo", "bar"]
}

resource "autodns_record" "example_MX" {
  zone_id = "foobar.test@bar.ns.net"

  name   = "foo"
  ttl    = 60
  type   = "MX"
  values = ["10 foo", "20 bar"]
}
