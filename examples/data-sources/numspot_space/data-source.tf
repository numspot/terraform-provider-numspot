resource "numspot_space" "test" {
  organisation_id = "88888888-4444-4444-4444-cccccccccccc"
  name            = "the space"
  description     = "the description"
}

data "numspot_space" "testdata" {
  space_id   = numspot_space.test.id
  depends_on = [numspot_space.test]
}