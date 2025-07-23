


//
// this chunk of code is used to determine the system theme preference and does not need to have the page loaded yet.
// This way the theme toggler can be initialized before the page is loaded and the theme can be set immediately.
class ThemeToggler{

    static root;
    static toggler;

    static themeTypes= ['light', 'dark', 'system', 'weather'];

    static systemTheme = 'light';
    static chosenTheme = 'system';
    static currentTheme = 'system';

    static init() {
        // set a reference to the root element
        ThemeToggler.root = document.documentElement;
        // Check localStorage for theme preference
        let themeLocalStorage = localStorage.getItem('theme');
        // if we found a theme in localStorage and it exists in the themetypes, set it to the currentTheme variable
        if (themeLocalStorage && ThemeToggler.themeTypes.includes(themeLocalStorage)) {
            //console.log('Theme found in localStorage:', themeLocalStorage);
            ThemeToggler.chosenTheme = themeLocalStorage;
        }
        // If the chosen theme is 'system', set the current theme to the system theme
        if(ThemeToggler.chosenTheme === 'system') {
            const prefersDarkScheme = window.matchMedia("(prefers-color-scheme: dark)");
            if (prefersDarkScheme.matches) {
                ThemeToggler.systemTheme = 'dark';
                ThemeToggler.currentTheme = 'dark';
            } else {
                ThemeToggler.systemTheme = 'light';
                ThemeToggler.currentTheme = 'light';
            }
            ThemeToggler.currentTheme = ThemeToggler.systemTheme;
        } else {
            // Otherwise, set the current theme to the chosen theme
            ThemeToggler.currentTheme = ThemeToggler.chosenTheme;
        }
        // Set the root element's data-theme attribute to the current theme
        //document.documentElement.className = ThemeToggler.currentTheme;
        ThemeToggler.root.dataset.theme = ThemeToggler.currentTheme;

    }

    static onLoad() {
        // Set the toggler reference
        ThemeToggler.toggler = document.getElementById('theme-toggle');
        // Add an event listener to the toggler
        ThemeToggler.toggler.addEventListener('click', ThemeToggler.toggleTheme);
    }

    static toggleTheme() {
        //console.log('Toggling theme...');
        // Toggle the theme between light and dark
        if (ThemeToggler.currentTheme === 'light') {
            ThemeToggler.currentTheme = 'dark';
        } else if (ThemeToggler.currentTheme === 'dark') {
            ThemeToggler.currentTheme = 'light';
        } else {
            // If the current theme is not light or dark, set it to light
            ThemeToggler.currentTheme = 'light';
        }
        // Set the root element's data-theme attribute to the current theme
        //document.documentElement.className = ThemeToggler.currentTheme;
        ThemeToggler.root.dataset.theme = ThemeToggler.currentTheme;
        // Save the theme to localStorage
        ThemeToggler.saveTheme(ThemeToggler.currentTheme);
    }

    static saveTheme(theme) {
        // Save the theme to localStorage
        localStorage.setItem('theme', ThemeToggler.currentTheme);
    }

}
// Initialize the ThemeToggler immediately (not waiting for the page to load)
ThemeToggler.init();


// Add an event listener to the window to call the onLoad method when the page is loaded
window.addEventListener('load',function(){
//window.onload = function() {

    // Call the onLoad method to set up the toggler and its event listeners
    ThemeToggler.onLoad();

    // Start fading out the loading overlay and start the progress bar
    setTimeout(() => {
        // fade out the loading overlay
        const loadingOverlay = document.getElementById('loading-overlay');
        if(loadingOverlay){
            loadingOverlay.style.transition = 'opacity 0.5s ease';
            loadingOverlay.style.opacity = '0';
            setTimeout(() => {
                loadingOverlay.style.display = 'none'; // hide the overlay after fading out
            }, 250); // match the timeout with the transition duration
        }
    }, 250);
    setInterval(() => {
        // Update the progress bar value
        const progressBar = document.getElementById('loading-overlay-progress');
        if (progressBar) {
            const value = Math.min(100, parseInt(progressBar.value) + 2);
            progressBar.value = value;
            if (value >= 100) {
                clearInterval(this); // stop the interval when it reaches 100
            }
        }
    }, 5); // update every 50ms/7 ... This feels like a good speed for the progress bar

});  
