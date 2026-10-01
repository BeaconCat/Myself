import { onMounted, onBeforeUnmount, ref } from 'vue';
const now = ref(Date.now());
let readers = 0, timer = 0;
const update = () => { now.value = Date.now(); };
/** Shared clock keeps relative labels current without one timer per card. */
export function useNow() {
  let registered = false;
  onMounted(() => {
    registered = true;
    update();
    if (readers++ === 0) {
      timer = window.setInterval(update, 30_000);
      document.addEventListener('visibilitychange', update);
    }
  });
  onBeforeUnmount(() => {
    if (!registered) return;
    if (--readers === 0) {
      window.clearInterval(timer);
      document.removeEventListener('visibilitychange', update);
    }
  });
  return now;
}
