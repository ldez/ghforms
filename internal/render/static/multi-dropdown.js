// Behavior for the multi-select dropdown rendered to look like regular single-line <select>.
// Classes used here are Tailwind utilities and must match the ones in _element.html / _layout.html.
(function () {
  const textGHMutedClass = 'text-gh-muted';
  const textGHFgClass = 'text-gh-fg';

  function getRequired(details) {
    return details.parentElement
      ? details.parentElement.querySelector(':scope > .multi-dropdown-required')
      : null;
  }

  function resetPlaceholder(placeholder) {
    placeholder.textContent = 'Select options';
    placeholder.classList.add(textGHMutedClass);
    placeholder.classList.remove(textGHFgClass);
  }

  function updateTrigger(details) {
    const placeholder = details.querySelector('.multi-dropdown-placeholder');
    const checks = details.querySelectorAll('[role="listbox"] input[type="checkbox"]');
    const required = getRequired(details);
    const selected = Array.from(checks).filter(c => c.checked).map(c => c.value);

    if (selected.length === 0) {
      resetPlaceholder(placeholder)
    } else {
      placeholder.textContent = selected.join(', ');
      placeholder.classList.remove(textGHMutedClass);
      placeholder.classList.add(textGHFgClass);
    }
    if (required) {
      required.value = selected.length > 0 ? selected.join(',') : '';
    }
  }

  function init(details) {
    details.addEventListener('change', function (e) {
      if (e.target && e.target.matches('input[type="checkbox"]')) {
        updateTrigger(details);
      }
    });
    updateTrigger(details);
  }

  document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('details.multi-dropdown').forEach(init);

    document.querySelectorAll('form').forEach(function (input) {
      input.addEventListener('reset', (event) => {
        document.querySelectorAll('.multi-dropdown-placeholder').forEach(resetPlaceholder);
      })
    });

    document.addEventListener('click', function (e) {
      document.querySelectorAll('details.multi-dropdown[open]').forEach(function (d) {
        if (!d.contains(e.target)) d.removeAttribute('open');
      });
    });
  });
})();
