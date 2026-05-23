/// Shodan endpoints. Obtain via [`crate::Client::shodan`].
pub struct Shodan<'a> {
    pub(crate) client: &'a crate::Client,
}
