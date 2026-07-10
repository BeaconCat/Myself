import type { Directive } from 'vue';

/** v-reveal：元素进入视口时播放入场动画（配合 styles/motion.scss 的 .reveal） */
export const vReveal: Directive<HTMLElement> = {
  mounted(el) {
    el.classList.add('reveal');
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            el.classList.add('reveal-in');
            observer.disconnect();
          }
        }
      },
      { threshold: 0.15 },
    );
    observer.observe(el);
  },
};
