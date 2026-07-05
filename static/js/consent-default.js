/*
 * Consent bootstrap — the single source of truth for consent configuration.
 *
 * Loaded as a BLOCKING <script> in <head>, before gtag/js, on every page. It
 * sets Google Consent Mode to denied-by-default and replays a stored, still-valid
 * "granted" choice before the GA library configures. consent.js reads the same
 * config from window.CONSENT_CONFIG, so the storage key / version / max-age /
 * GA id live in exactly one place.
 */
(function (w) {
  'use strict';

  var CONFIG = {
    storageKey: 'cookie-consent',
    version: 1,                          // bump to force re-consent everywhere
    maxAgeMs: 1000 * 60 * 60 * 24 * 365, // ~12 months, then re-prompt
    gaId: 'G-RH8KWHKMPZ'
  };
  w.CONSENT_CONFIG = CONFIG;

  w.dataLayer = w.dataLayer || [];
  function gtag() { w.dataLayer.push(arguments); }
  w.gtag = gtag;

  // Deny everything by default, before the GA library loads.
  gtag('consent', 'default', {
    'analytics_storage': 'denied',
    'ad_storage': 'denied',
    'ad_user_data': 'denied',
    'ad_personalization': 'denied'
  });

  // Replay a previously stored, still-valid "granted" choice.
  try {
    var raw = w.localStorage.getItem(CONFIG.storageKey);
    if (raw) {
      var c = JSON.parse(raw);
      if (c && c.version === CONFIG.version && c.status === 'granted' &&
          typeof c.timestamp === 'number' && (Date.now() - c.timestamp) < CONFIG.maxAgeMs) {
        gtag('consent', 'update', { 'analytics_storage': 'granted' });
      }
    }
  } catch (e) { /* storage blocked → stay denied (the safe default) */ }

  gtag('js', new Date());
  gtag('config', CONFIG.gaId);
})(window);
