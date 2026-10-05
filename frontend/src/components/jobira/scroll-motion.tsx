import { useEffect } from "react";

const TARGETS = "main section, main article, main [data-reveal]";

export function ScrollMotion() {
  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("is-visible");
            observer.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.08, rootMargin: "0px 0px -48px" },
    );

    const register = () => {
      document.querySelectorAll(TARGETS).forEach((element, index) => {
        if (element.classList.contains("scroll-reveal") || element.classList.contains("is-visible")) return;
        element.classList.add("scroll-reveal");
        if (index < 8) element.classList.add(`reveal-delay-${index % 4}`);
        observer.observe(element);
      });
    };

    let mutations: MutationObserver | undefined;
    const timer = window.setTimeout(() => {
      register();
      mutations = new MutationObserver(register);
      mutations.observe(document.body, { childList: true, subtree: true });
    }, 120);
    return () => {
      window.clearTimeout(timer);
      mutations?.disconnect();
      observer.disconnect();
    };
  }, []);

  return null;
}