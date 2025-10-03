/**
 * KubeStellar Documentation Search Keyboard Shortcut
 * 
 * This script adds a keyboard shortcut (Ctrl+K / Cmd+K) to focus the search input
 * across all platforms (Windows, macOS, Linux).
 * 
 * The shortcut works by:
 * 1. Listening for the keydown event
 * 2. Checking for Ctrl+K (Windows/Linux) or Cmd+K (macOS)
 * 3. Focusing the search input and opening the search dialog
 * 4. Preventing default browser behavior
 */

(function() {
    'use strict';

    // Configuration
    const SEARCH_INPUT_SELECTOR = 'input[data-md-component="search-query"]';
    const SEARCH_TOGGLE_SELECTOR = 'input[data-md-toggle="search"]';
    const SEARCH_OVERLAY_SELECTOR = '.md-search__overlay';
    const SEARCH_ICON_SELECTOR = 'label[for="__search"]';
    
    // Key codes and modifiers
    const KEY_K = 'k';
    const KEY_K_CODE = 75;
    
    // Track if we're already in search mode to avoid conflicts
    let isSearchFocused = false;
    let hintElement = null;
    let hintTimeout = null;
    
    /**
     * Create and show the keyboard shortcut hint
     */
    function showHint() {
        if (hintElement) {
            clearTimeout(hintTimeout);
            hintElement.classList.add('md-search__hint--visible');
            return;
        }
        
        const searchIcon = document.querySelector(SEARCH_ICON_SELECTOR);
        if (!searchIcon) return;
        
        // Create hint element
        hintElement = document.createElement('div');
        hintElement.className = 'md-search__hint';
        hintElement.setAttribute('role', 'tooltip');
        hintElement.setAttribute('aria-live', 'polite');
        hintElement.innerHTML = 'Press <kbd>Ctrl+K</kbd> to search';
        
        // Detect platform for appropriate shortcut display
        const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
        if (isMac) {
            hintElement.innerHTML = 'Press <kbd>⌘K</kbd> to search';
        }
        
        // Position relative to search icon
        searchIcon.style.position = 'relative';
        searchIcon.appendChild(hintElement);
        
        // Show with animation
        setTimeout(() => {
            if (hintElement) {
                hintElement.classList.add('md-search__hint--visible');
            }
        }, 10);
        
        // Auto-hide after 3 seconds
        hintTimeout = setTimeout(() => {
            hideHint();
        }, 3000);
    }
    
    /**
     * Hide the keyboard shortcut hint
     */
    function hideHint() {
        if (hintElement) {
            hintElement.classList.remove('md-search__hint--visible');
            clearTimeout(hintTimeout);
            hintTimeout = null;
        }
    }
    
    /**
     * Focus the search input and open the search dialog
     */
    function focusSearch() {
        try {
            // Find the search input element
            const searchInput = document.querySelector(SEARCH_INPUT_SELECTOR);
            const searchToggle = document.querySelector(SEARCH_TOGGLE_SELECTOR);
            
            if (!searchInput || !searchToggle) {
                console.warn('KubeStellar Search Shortcut: Search elements not found');
                return;
            }
            
            // Open the search dialog by checking the toggle
            if (!searchToggle.checked) {
                searchToggle.checked = true;
                // Trigger change event to ensure Material theme responds
                searchToggle.dispatchEvent(new Event('change', { bubbles: true }));
            }
            
            // Small delay to ensure the search dialog is fully opened
            setTimeout(() => {
                searchInput.focus();
                searchInput.select(); // Select any existing text for easy replacement
                isSearchFocused = true;
            }, 50);
            
        } catch (error) {
            console.error('KubeStellar Search Shortcut: Error focusing search', error);
        }
    }
    
    /**
     * Handle keyboard events
     */
    function handleKeyDown(event) {
        // Check if K key is pressed
        const isKKey = event.key === KEY_K || event.keyCode === KEY_K_CODE;
        
        // Check for Ctrl (Windows/Linux) or Cmd (macOS)
        const isModifierPressed = (event.ctrlKey && !event.metaKey) || (!event.ctrlKey && event.metaKey);
        
        // Check if we're not already in a text input/textarea/contenteditable
        const isInInput = event.target.tagName === 'INPUT' || 
                         event.target.tagName === 'TEXTAREA' || 
                         event.target.contentEditable === 'true';
        
        // Check if we're in the search input specifically
        const isInSearchInput = event.target.matches(SEARCH_INPUT_SELECTOR);
        
        // Only trigger if:
        // 1. K key is pressed
        // 2. Modifier (Ctrl/Cmd) is pressed
        // 3. We're not in a regular input field (unless it's the search input)
        // 4. We're not already focused on search
        if (isKKey && isModifierPressed && (!isInInput || isInSearchInput) && !isSearchFocused) {
            event.preventDefault();
            event.stopPropagation();
            focusSearch();
        }
        
        // Show hint when user starts typing Ctrl/Cmd (but not K yet)
        if (isModifierPressed && !isKKey && !isInInput) {
            showHint();
        }
    }
    
    /**
     * Clear the search input
     */
    function clearSearchInput() {
        const searchInput = document.querySelector(SEARCH_INPUT_SELECTOR);
        if (searchInput) {
            searchInput.value = '';
        }
    }
    
    /**
     * Handle when search loses focus
     */
    function handleSearchBlur() {
        isSearchFocused = false;
    }
    
    /**
     * Handle escape key to close search
     */
    function handleKeyUp(event) {
        if (event.key === 'Escape' && isSearchFocused) {
            const searchToggle = document.querySelector(SEARCH_TOGGLE_SELECTOR);
            if (searchToggle && searchToggle.checked) {
                searchToggle.checked = false;
                searchToggle.dispatchEvent(new Event('change', { bubbles: true }));
                isSearchFocused = false;
                // Clear the search input when closing
                clearSearchInput();
            }
        }
    }
    
    /**
     * Handle mouse hover on search icon
     */
    function handleSearchHover() {
        showHint();
    }
    
    /**
     * Handle mouse leave on search icon
     */
    function handleSearchLeave() {
        hideHint();
    }
    
    /**
     * Initialize the search shortcut functionality
     */
    function init() {
        // Wait for DOM to be ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', init);
            return;
        }
        
        // Add event listeners
        document.addEventListener('keydown', handleKeyDown, true);
        document.addEventListener('keyup', handleKeyUp, true);
        
        // Listen for search input blur events
        const searchInput = document.querySelector(SEARCH_INPUT_SELECTOR);
        if (searchInput) {
            searchInput.addEventListener('blur', handleSearchBlur);
            // Set custom placeholder text
            searchInput.placeholder = 'Cmd/Ctrl+K';
        }
        
        // Listen for search dialog close events
        const searchToggle = document.querySelector(SEARCH_TOGGLE_SELECTOR);
        if (searchToggle) {
            searchToggle.addEventListener('change', function() {
                if (!this.checked) {
                    isSearchFocused = false;
                    // Clear the search input when modal is closed
                    clearSearchInput();
                }
            });
        }
        
        // Add hover events to search icon
        const searchIcon = document.querySelector(SEARCH_ICON_SELECTOR);
        if (searchIcon) {
            searchIcon.addEventListener('mouseenter', handleSearchHover);
            searchIcon.addEventListener('mouseleave', handleSearchLeave);
        }
        
        // Hide hint when clicking anywhere
        document.addEventListener('click', function(event) {
            if (!event.target.closest('.md-search__hint') && !event.target.closest(SEARCH_ICON_SELECTOR)) {
                hideHint();
            }
        });
        
        console.log('KubeStellar Search Shortcut: Initialized (Ctrl+K / Cmd+K)');
    }
    
    // Initialize when script loads
    init();
    
})();
