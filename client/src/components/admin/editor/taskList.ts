import TaskList from '@tiptap/extension-task-list';
import type MarkdownIt from 'markdown-it';
// @ts-expect-error The offline plugin does not ship types.
import taskLists from 'markdown-it-task-lists';
const configured = new WeakSet<MarkdownIt>();

/** Tiptap has distinct list node types; split mixed GFM lists without adding empty tasks. */
export const MarkdownTaskList = TaskList.extend({
  addStorage() {
    return { markdown: { parse: {
      setup(md: MarkdownIt) {
        if (configured.has(md)) return;
        configured.add(md); md.use(taskLists);
      },
      updateDOM(element: HTMLElement) {
        for (const list of [...element.querySelectorAll('ul.contains-task-list')].reverse()) {
          const fragment = document.createDocumentFragment();
          let group: Element | null = null, previous: boolean | null = null;
          for (const item of [...list.children]) {
            const task = item.classList.contains('task-list-item');
            if (task !== previous) {
              group = list.cloneNode(false) as Element;
              if (task) group.setAttribute('data-type', 'taskList');
              else { group.removeAttribute('data-type'); group.classList.remove('contains-task-list'); }
              fragment.append(group); previous = task;
            }
            group!.append(item);
          }
          list.replaceWith(fragment);
        }
      },
    } } };
  },
});
