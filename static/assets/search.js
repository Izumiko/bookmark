(() => {
  // Cache DOM elements
  const searchForm = document.getElementById('search-form');
  const searchInput = document.getElementById('search-input');
  const searchStyle = document.getElementById('search-style');
  
  // Add event listeners
  document.addEventListener('keyup', handleGlobalKeyup);
  searchForm.addEventListener('submit', handleSearchFormSubmit);
  searchInput.addEventListener('input', debounce(handleSearchInputKeyup, 150));
  searchInput.addEventListener('blur', handleSearchInputBlur);
  
  // Debounce function to limit rapid firing of events
  function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  // Handle global keyboard shortcuts
  function handleGlobalKeyup(e) {
    if (e.altKey || e.ctrlKey || e.metaKey || document.activeElement === searchInput) {
      return;
    }
    
    const key = e.key.toLowerCase();
    if (/^[a-z0-9]$/.test(key)) {
      searchInput.focus();
      searchInput.value = key;
      handleSearchInputKeyup();
    }
  }

  // Handle form submission
  function handleSearchFormSubmit(e) {
    e.preventDefault();
    const focusedResult = document.querySelector('.site-bookmark-a-focus');
    if (focusedResult) {
      focusedResult.click();
    }
  }

  // Handle search input events
  function handleSearchInputKeyup() {
    clearTabindex();
    clearFocus();

    const query = searchInput.value.trim().toLowerCase();
    if (query === '') {
      searchStyle.innerHTML = '';
      return;
    }

    const splitedQuery = query.split(/\s+/);
    
    setSearchStyle(splitedQuery);
    setTabindex(splitedQuery);
    setFocus(splitedQuery);
  }

  // Handle input blur event
  function handleSearchInputBlur() {
    const focusedResult = document.querySelector('.site-bookmark-a-focus');
    if (focusedResult) {
      focusedResult.classList.remove('site-bookmark-a-focus');
    }
  }

  // Reset tabindex to default for all elements
  function clearTabindex() {
    document.querySelectorAll('[tabindex="2"]').forEach(element => {
      element.setAttribute('tabindex', '9');
    });
  }

  // Clear focus styling from all elements
  function clearFocus() {
    document.querySelectorAll('.site-bookmark-a-focus').forEach(element => {
      element.classList.remove('site-bookmark-a-focus');
    });
  }

  // Set CSS for search filtering
  function setSearchStyle(splitedQuery) {
    // Use template literals for better readability
    searchStyle.innerHTML = `
      .site-bookmark-category {
        display: none;
      }
      .site-bookmark-a {
        opacity: 0.3;
      }
      ${generateQuerySelectorQuery(splitedQuery)} {
        order: -1;
        -ms-flex-order: -1;
      }
      ${generateQuerySelectorQuery(splitedQuery)} .site-bookmark-a {
        opacity: 1;
      }
    `;
  }

  // Update tabindex for search results
  function setTabindex(splitedQuery) {
    const selector = `${generateQuerySelectorQuery(splitedQuery)} .site-bookmark-a`;
    document.querySelectorAll(selector).forEach(item => {
      item.setAttribute('tabindex', '2');
    });
  }

  // Set focus on first matching result
  function setFocus(splitedQuery) {
    const firstItem = document.querySelector(`${generateQuerySelectorQuery(splitedQuery)} .site-bookmark-a`);
    if (firstItem) {
      firstItem.classList.add('site-bookmark-a-focus');
    }
  }

  // Create CSS selector from search terms
  function generateQuerySelectorQuery(splitedQuery) {
    return splitedQuery.map(query => `[data-name*="${query}"]`).join('');
  }
})();