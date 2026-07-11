// Small progressive-enhancement helpers for Smart 360.
//
// Kept CSP-friendly: this is a self-hosted script (script-src 'self'), so there
// are no inline scripts or event handlers anywhere in the templates.
(function () {
  "use strict";

  // Navigate when a swapped-in fragment asks for it via [data-redirect].
  // Used by the SSE consolidation flow: the "done" event swaps in an element
  // carrying data-redirect, and we send the browser there — no inline script.
  function handleRedirect() {
    var el = document.querySelector("[data-redirect]");
    if (el) {
      var url = el.getAttribute("data-redirect");
      el.removeAttribute("data-redirect"); // guard against repeat navigation
      if (url) {
        window.location.assign(url);
      }
    }
  }

  // First-login onboarding tour: pure client-side slide navigation. Skip/finish
  // are htmx POSTs to /onboarding/complete (which removes the overlay); this only
  // handles Next/Back and the slide/dot/button visibility.
  function initOnboarding() {
    var root = document.getElementById("onboarding");
    if (!root) return;

    var slides = root.querySelectorAll(".onboarding__slide");
    var dots = root.querySelectorAll(".onboarding__dots span");
    var prevBtn = root.querySelector('[data-tour="prev"]');
    var nextBtn = root.querySelector('[data-tour="next"]');
    var finishBtn = root.querySelector('[data-tour="finish"]');
    var current = 0;

    function render() {
      slides.forEach(function (s, i) { s.classList.toggle("is-active", i === current); });
      dots.forEach(function (d, i) { d.classList.toggle("is-active", i === current); });
      var last = current === slides.length - 1;
      if (prevBtn) prevBtn.hidden = current === 0;
      if (nextBtn) nextBtn.hidden = last;
      if (finishBtn) finishBtn.hidden = !last;
    }
    if (nextBtn) nextBtn.addEventListener("click", function () {
      if (current < slides.length - 1) { current++; render(); }
    });
    if (prevBtn) prevBtn.addEventListener("click", function () {
      if (current > 0) { current--; render(); }
    });
    render();
  }

  document.addEventListener("DOMContentLoaded", function () {
    document.body.addEventListener("htmx:afterSwap", handleRedirect);
    document.body.addEventListener("htmx:oobAfterSwap", handleRedirect);
    initOnboarding();
  });
})();
