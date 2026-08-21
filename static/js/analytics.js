/*
 * Analytics bootstrap: consent state + the one event helper.
 *
 * These live in the same file on purpose. The event layer and the consent layer
 * share one lifecycle and one configuration, and splitting them invites a second
 * source of truth for the tag.
 *
 * ---------------------------------------------------------------------------
 * CONSENT
 *
 * Owns the visitor's analytics choice and the consent bar. Configuration
 * (storage key / version / max-age) comes from window.CONSENT_CONFIG, set by the
 * inline bootstrap in <head> that also seeds Consent Mode before the tag loads —
 * so those values live in exactly one place. The fallback below mirrors the
 * server defaults and only applies if the bootstrap was stripped.
 *
 * Exposes window.cookieConsent:
 *   getChoice()      -> 'granted' | 'denied' | null   (null = no valid choice)
 *   setChoice(status)-> persists + updates Consent Mode
 *   shouldShow()     -> true when the bar should be displayed
 *   onChange(fn)     -> subscribe to choice changes (bar uses this to hide)
 *
 * ---------------------------------------------------------------------------
 * TRACKING
 *
 * One track(name, params) helper, exposed as window.track. It branches on the
 * tag kind the SERVER already resolved (window.__tag.kind, rendered from
 * site.Tag) — never on `typeof gtag`, and never by prefix-matching an id here.
 * A bare dataLayer.push on the GA4 branch records nothing, and a gtag('event')
 * on the GTM branch bypasses the container, so getting this wrong is the silent
 * no-data failure the whole design exists to prevent.
 *
 *   kind 'ga4' -> gtag('event', name, params)
 *   kind 'gtm' -> dataLayer.push({event: name, ...params})
 *   no tag     -> silent no-op (local dev and previews must never throw)
 *
 * Wiring is declarative and centralised: one delegated document-level listener
 * reads data-track / data-track-* off the element actually used. No inline
 * onclick, no per-component listeners, no analytics snippets in templates.
 *
 * Events fire under Consent Mode's default-denied state too — GA runs cookieless
 * until granted. There is deliberately no local queue replaying actions after
 * consent.
 *
 * NEVER put PII in a param: no email, phone, name or message text. Params stay
 * low-cardinality (ids, canonical labels, locale, section).
 */
(function (window) {
  'use strict';

  var CFG = window.CONSENT_CONFIG || {
    storageKey: 'cookie-consent',
    version: 1,
    maxAgeMs: 1000 * 60 * 60 * 24 * 180   // ~6 months, then re-prompt
  };
  var STORAGE_KEY = CFG.storageKey;
  var VERSION = CFG.version;
  var MAX_AGE = CFG.maxAgeMs;

  var listeners = [];

  // ---- Event vocabulary -----------------------------------------------------
  // One canonical name per action, defined once. GA4's recommended names are
  // used verbatim where one exists; two spellings of an event is two
  // half-populated reports.
  var EVENTS = {
    fileDownload:  'file_download',
    contactClick:  'contact_click',
    selectContent: 'select_content',
    themeToggle:   'theme_toggle',
    consentUpdate: 'consent_update',
    pageNotFound:  'page_not_found'
  };

  // ---- track ---------------------------------------------------------------
  function track(name, params) {
    if (!name) return;
    var tag = window.__tag;
    if (!tag || !tag.kind) return;              // no tag configured -> no-op

    var payload = params || {};
    try {
      if (tag.kind === 'ga4') {
        if (typeof window.gtag === 'function') {
          window.gtag('event', name, payload);
        }
      } else if (tag.kind === 'gtm') {
        window.dataLayer = window.dataLayer || [];
        var ev = { event: name };
        for (var k in payload) {
          if (Object.prototype.hasOwnProperty.call(payload, k)) ev[k] = payload[k];
        }
        window.dataLayer.push(ev);
      }
    } catch (e) {
      // Analytics must never break the page it measures.
    }
  }

  // Reads data-track-* attributes off an element into an event param object.
  // "data-track-file-name" becomes "file_name"; the literal value "$path"
  // resolves to the current path, which is how the 404 reports which URL missed
  // (its body is one pre-rendered blob shared by every unmatched path, so the
  // server cannot bake the path in).
  function paramsFrom(el) {
    var out = {};
    var attrs = el.attributes;
    for (var i = 0; i < attrs.length; i++) {
      var name = attrs[i].name;
      if (name.indexOf('data-track-') !== 0) continue;
      var key = name.slice('data-track-'.length).replace(/-/g, '_');
      if (key === 'load' || key === '') continue;   // control attribute, not a param
      var value = attrs[i].value;
      out[key] = (value === '$path') ? window.location.pathname : value;
    }
    return out;
  }

  // One delegated listener for every click-triggered event on the site.
  function initTracking(doc) {
    doc.addEventListener('click', function (e) {
      var target = e.target;
      if (!target || !target.closest) return;
      var el = target.closest('[data-track]');
      if (!el) return;
      track(el.getAttribute('data-track'), paramsFrom(el));
    });

    // On-load events (currently only the 404) are declared the same way, with
    // data-track-load instead of data-track.
    var onLoad = doc.querySelectorAll('[data-track-load]');
    for (var i = 0; i < onLoad.length; i++) {
      track(onLoad[i].getAttribute('data-track-load'), paramsFrom(onLoad[i]));
    }
  }

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

    // Fires on every recorded choice, including Esc. Under default-denied GA
    // runs cookieless, so this still reaches the tag.
    track(EVENTS.consentUpdate, { status: status });

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

  // The single entry point for every event on the site. theme.js calls it too,
  // guarding for absence so the theme control never depends on analytics.
  window.track = track;
  window.trackEvents = EVENTS;

  // ---- Consent bar controller ---------------------------------------------
  // Reveals the bar only when there's no valid stored choice. Button click
  // wiring is added in the next task; hide-on-choice is handled via onChange.
  function initBar() {
    var doc = window.document;
    if (!doc) return;
    var bar = doc.getElementById('cookie-bar');
    if (!bar) return;

    // The bar is fixed-position, so the page must reserve its height or it
    // covers the footer. Measured rather than hardcoded: the bar wraps to two
    // rows at narrow widths, so its height is not a constant.
    function reserveSpace() {
      var h = bar.getBoundingClientRect().height;
      doc.documentElement.style.setProperty('--consent-bar-height', Math.ceil(h) + 'px');
    }

    function releaseSpace() {
      doc.documentElement.style.removeProperty('--consent-bar-height');
    }

    function reveal(focusFirst) {
      bar.hidden = false;
      // next frame so the entrance animation plays from the hidden state
      (window.requestAnimationFrame || function (f) { f(); })(function () {
        bar.classList.add('is-visible');
        reserveSpace();
        if (focusFirst) {
          var first = bar.querySelector('[data-consent]');
          if (first && first.focus) first.focus();
        }
      });
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

    // Esc DECLINES (only while the bar is open).
    //
    // It records a real 'denied' choice rather than dismissing silently. Both
    // are privacy-safe — consent stays denied either way — but dismissing
    // without recording stored nothing, so the bar came back on every single
    // visit forever for anyone who closed it with the keyboard. Escaping a
    // non-modal bar is a decision, so we honour it as one.
    doc.addEventListener('keydown', function (e) {
      if ((e.key === 'Escape' || e.keyCode === 27) && !bar.hidden) {
        setChoice('denied');
      }
    });

    // First visit (no valid stored choice) → show automatically, without
    // stealing focus from the page (the bar is non-modal).
    if (shouldShow()) reveal(false);

    // Once a choice is made, retire the bar and give the space back.
    onChange(function () {
      bar.classList.remove('is-visible');
      bar.hidden = true;
      releaseSpace();
    });

    // Keep the reservation correct when the bar reflows (rotation, resize).
    window.addEventListener('resize', function () {
      if (!bar.hidden) reserveSpace();
    });
  }

  if (window.document) {
    // Tracking is wired even when no consent bar exists: with no tag configured
    // every track() call is a no-op, and the delegated listener costs nothing.
    var start = function () {
      initTracking(window.document);
      initBar();
    };
    if (window.document.readyState !== 'loading') {
      start();
    } else {
      window.document.addEventListener('DOMContentLoaded', start);
    }
  }
})(window);
