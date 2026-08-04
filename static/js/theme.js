/*
 * Theme toggle — cycles dark → light → auto.
 *
 * `auto` follows prefers-color-scheme and is the state before any choice is
 * made; the two explicit modes override it by setting data-theme on <html>.
 * The chosen mode is persisted in localStorage as the literal string
 * "dark" | "light" | "auto".
 *
 * The initial attribute is applied pre-paint by an inline <head> script (see
 * templates/partials/head.html) so the correct theme is painted on the first
 * frame. This file only owns the control.
 */
(function (window) {
  'use strict';

  var STORAGE_KEY = 'theme';
  var MODES = ['dark', 'light', 'auto'];
  var LABELS = { dark: 'Dark', light: 'Light', auto: 'Auto' };

  var doc = window.document;
  if (!doc) return;

  function read() {
    try {
      var stored = window.localStorage.getItem(STORAGE_KEY);
      return MODES.indexOf(stored) > -1 ? stored : 'auto';
    } catch (e) {
      return 'auto';                                 // storage blocked
    }
  }

  function apply(mode) {
    if (mode === 'auto') {
      doc.documentElement.removeAttribute('data-theme');
    } else {
      doc.documentElement.setAttribute('data-theme', mode);
    }
    try {
      window.localStorage.setItem(STORAGE_KEY, mode);
    } catch (e) {
      // Private mode / storage blocked — the choice still applies this session.
    }
  }

  function initToggle() {
    var button = doc.querySelector('[data-theme-toggle]');
    if (!button) return;

    var labelEl = button.querySelector('[data-theme-label]');
    var mode = read();

    function render() {
      if (labelEl) labelEl.textContent = LABELS[mode];
      // The visible label is the non-colour state cue; the accessible name
      // spells out both the current state and what pressing will do.
      var next = MODES[(MODES.indexOf(mode) + 1) % MODES.length];
      button.setAttribute(
        'aria-label',
        'Theme: ' + LABELS[mode] + '. Switch to ' + LABELS[next] + '.'
      );
    }

    button.addEventListener('click', function () {
      mode = MODES[(MODES.indexOf(mode) + 1) % MODES.length];
      apply(mode);
      render();
    });

    button.hidden = false;   // only reveal the control once it can work
    render();
  }

  if (doc.readyState !== 'loading') {
    initToggle();
  } else {
    doc.addEventListener('DOMContentLoaded', initToggle);
  }
})(window);
