/// ULP endpoints. Obtain via [`crate::Client::ulp`].
pub struct Ulp<'a> {
    pub(crate) client: &'a crate::Client,
}
