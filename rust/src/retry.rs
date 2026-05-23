use std::time::Duration;

/// Controls automatic retries. Disabled by default (`max_retries == 0`).
#[derive(Debug, Clone)]
pub struct RetryConfig {
    pub max_retries: u32,
    pub base_delay: Duration,
    pub max_delay: Duration,
    pub retry_on: Vec<u16>,
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_retries: 0,
            base_delay: Duration::from_millis(500),
            max_delay: Duration::from_secs(8),
            retry_on: vec![429, 500, 502, 503, 504],
        }
    }
}

/// Full-jitter exponential backoff. `rnd` must be in `[0, 1)`.
/// When `retry_after` is set it takes precedence, capped at `max_delay`.
pub(crate) fn backoff_delay(
    attempt: u32,
    cfg: &RetryConfig,
    retry_after: Option<Duration>,
    rnd: f64,
) -> Duration {
    if let Some(ra) = retry_after {
        return ra.min(cfg.max_delay);
    }
    let factor = 1u64.checked_shl(attempt).unwrap_or(u64::MAX);
    let exp = cfg
        .base_delay
        .checked_mul(factor as u32)
        .unwrap_or(cfg.max_delay)
        .min(cfg.max_delay);
    exp.mul_f64(rnd)
}

pub(crate) fn should_retry(status: u16, attempt: u32, cfg: &RetryConfig) -> bool {
    attempt < cfg.max_retries && cfg.retry_on.contains(&status)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn cfg() -> RetryConfig {
        RetryConfig {
            max_retries: 3,
            base_delay: Duration::from_millis(500),
            max_delay: Duration::from_secs(8),
            retry_on: vec![429, 500, 502, 503, 504],
        }
    }

    #[test]
    fn retry_after_wins_capped() {
        assert_eq!(
            backoff_delay(0, &cfg(), Some(Duration::from_secs(1)), 0.5),
            Duration::from_secs(1)
        );
        assert_eq!(
            backoff_delay(0, &cfg(), Some(Duration::from_secs(3600)), 0.5),
            Duration::from_secs(8)
        );
    }

    #[test]
    fn full_jitter_on_exponential() {
        assert_eq!(
            backoff_delay(0, &cfg(), None, 0.5),
            Duration::from_millis(250)
        );
        assert_eq!(backoff_delay(2, &cfg(), None, 0.5), Duration::from_secs(1));
    }

    #[test]
    fn never_exceeds_max_delay() {
        assert!(backoff_delay(40, &cfg(), None, 1.0) <= Duration::from_secs(8));
    }

    #[test]
    fn should_retry_rules() {
        assert!(should_retry(429, 0, &cfg()));
        assert!(should_retry(503, 2, &cfg()));
        assert!(!should_retry(429, 3, &cfg()));
        assert!(!should_retry(400, 0, &cfg()));
    }

    #[test]
    fn default_is_disabled() {
        assert_eq!(RetryConfig::default().max_retries, 0);
    }
}
