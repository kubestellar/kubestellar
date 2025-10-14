document.addEventListener("DOMContentLoaded", function() {
    const bar = document.querySelector(".custom-announcement");
    if (!bar) return;
  
    let lastScrollY = window.scrollY;
    const threshold = 150; // adjust this as needed (px)
  
    window.addEventListener("scroll", () => {
      if (window.scrollY > threshold) {
        bar.classList.add("hidden");
      } else {
        bar.classList.remove("hidden");
      }
      lastScrollY = window.scrollY;
    });
  });
