/*
 * Cookie consent state manager.
 *
 * Owns the visitor's analytics choice and the consent bar. Configuration (storage
 * key / version / max-age) comes from window.CONSENT_CONFIG, set by the shared
 * /static/js/consent-default.js bootstrap that also seeds Consent Mode in <head> —
 * so those values live in exactly one place. The fallback below only applies if
 * that bootstrap failed to load.
 *
 * Exposes window.cookieConsent:
 *   getChoice()      -> 'granted' | 'denied' | null   (null = no valid choice)
 *   setChoice(status)-> persists + updates Consent Mode
 *   shouldShow()     -> true when the bar should be displayed
 *   onChange(fn)     -> subscribe to choice changes (bar uses this to hide)
 */
(function (window) {
  'use strict';

  var CFG = window.CONSENT_CONFIG || {
    storageKey: 'cookie-consent',
    version: 1,
    maxAgeMs: 1000 * 60 * 60 * 24 * 365
  };
  var STORAGE_KEY = CFG.storageKey;
  var VERSION = CFG.version;
  var MAX_AGE = CFG.maxAgeMs;

  var listeners = [];

  function gtag() {
    // gtag is defined by the inline head snippet; guard in case it is absent.
    if (typeof window.gtag === 'function') {
      window.gtag.apply(window, arguments);
    }
  }

  function read() {
    try {
      var raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      var c = JSON.parse(raw);
      if (!c || c.version !== VERSION) return null;
      if (typeof c.timestamp !== 'number') return null;
      if ((Date.now() - c.timestamp) >= MAX_AGE) return null;   // expired
      if (c.status !== 'granted' && c.status !== 'denied') return null;
      return c;
    } catch (e) {
      return null;                                   // storage blocked/corrupt
    }
  }

  function getChoice() {
    var c = read();
    return c ? c.status : null;
  }

  function shouldShow() {
    return getChoice() === null;
  }

  function setChoice(status) {
    if (status !== 'granted' && status !== 'denied') return;

    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify({
        status: status,
        version: VERSION,
        timestamp: Date.now()
      }));
    } catch (e) {
      // Persisting failed (private mode / blocked). Still apply for this session.
    }

    gtag('consent', 'update', {
      'analytics_storage': status === 'granted' ? 'granted' : 'denied'
    });

    for (var i = 0; i < listeners.length; i++) {
      try { listeners[i](status); } catch (e) { /* ignore listener errors */ }
    }
  }

  function onChange(fn) {
    if (typeof fn === 'function') listeners.push(fn);
  }

  window.cookieConsent = {
    getChoice: getChoice,
    shouldShow: shouldShow,
    setChoice: setChoice,
    onChange: onChange
  };

  // ---- Consent bar controller ---------------------------------------------
  // Reveals the bar only when there's no valid stored choice. Button click
  // wiring is added in the next task; hide-on-choice is handled via onChange.
  function initBar() {
    var doc = window.document;
    if (!doc) return;
    var bar = doc.getElementById('cookie-bar');
    if (!bar) return;

    function reveal(focusFirst) {
      bar.hidden = false;
      // next frame so the entrance animation plays from the hidden state
      (window.requestAnimationFrame || function (f) { f(); })(function () {
        bar.classList.add('is-visible');
        if (focusFirst) {
          var first = bar.querySelector('[data-consent]');
          if (first && first.focus) first.focus();
        }
      });
    }

    // Dismiss the bar WITHOUT recording a choice. Consent stays denied (the safe
    // default) and the bar simply returns on the next visit — Esc never grants.
    function dismissNoChoice() {
      bar.classList.remove('is-visible');
      bar.hidden = true;
    }

    // Wire the Accept / Decline buttons: each carries data-consent.
    var buttons = bar.querySelectorAll('[data-consent]');
    for (var i = 0; i < buttons.length; i++) {
      buttons[i].addEventListener('click', function () {
        setChoice(this.getAttribute('data-consent'));
      });
    }

    // Wire any "Cookie settings" trigger to reopen the bar so a visitor can
    // withdraw/change consent as easily as they gave it (there is at most one
    // per page, but support several defensively).
    var reopeners = doc.querySelectorAll('[data-consent-reopen]');
    for (var j = 0; j < reopeners.length; j++) {
      reopeners[j].addEventListener('click', function () { reveal(true); });
    }

    // Esc dismisses the bar without consenting (only while it's open).
    doc.addEventListener('keydown', function (e) {
      if ((e.key === 'Escape' || e.keyCode === 27) && !bar.hidden) {
        dismissNoChoice();
      }
    });

    // First visit (no valid stored choice) → show automatically, without
    // stealing focus from the page (the bar is non-modal).
    if (shouldShow()) reveal(false);

    // Once a choice is made, retire the bar.
    onChange(function () {
      bar.classList.remove('is-visible');
      bar.hidden = true;
    });
  }

  if (window.document) {
    if (window.document.readyState !== 'loading') {
      initBar();
    } else {
      window.document.addEventListener('DOMContentLoaded', initBar);
    }
  }
})(window);
