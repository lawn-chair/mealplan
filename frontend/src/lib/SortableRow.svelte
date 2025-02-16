<script lang="ts">
    import type { Snippet } from "svelte";
    import type { HTMLAttributes } from "svelte/elements";
  
    const DRAG_CLASS = "dragging";
    const DRAG_OVER_CLASS = "dragging-over";
  
    type Props = {
      children: Snippet;
      drag?: (from: number, to: number) => void;
      dragClass?: string;
      dragGhostClass?: string;
      dragOverClass?: string;
      handleClass?: string;
      this: keyof HTMLElementTagNameMap;
    } & HTMLAttributes<never>;
 
    let {
      children,
      drag,
      dragClass: dragClassProp,
      dragGhostClass: dragGhostClassProp,
      dragOverClass: dragOverClassProp,
      handleClass: handleClassProp,
      this: elementTag,
      ...htmlAttributes
    }: Props = $props();
  
    let itemEl: HTMLElement | undefined = $state();
    let parentEl = $derived(itemEl?.parentElement);
    let handleClass = $derived(handleClassProp ? `.${handleClassProp}` : undefined);
    let handle = $derived(handleClass ? itemEl?.querySelector(handleClass) : undefined);
    let dragClass = $derived((dragClassProp || "bg-gray-200").split(" ").filter((c) => c));
    let dragGhostClass = $derived((dragGhostClassProp || "bg-yellow-300").split(" ").filter((c) => c));
    let dragOverClass = $derived((dragOverClassProp || "bg-blue-300").split(" ").filter((c) => c));
  
    const getItemFromTarget = (target: HTMLElement, prevChildElement?: HTMLElement): HTMLElement => {
      if (target === parentEl) return prevChildElement as HTMLElement;
  
      return getItemFromTarget(target.parentElement!, target) as HTMLElement;
    };
  
    const getItemIndex = (item: HTMLElement) =>
      Array.from(parentEl!.querySelectorAll(`:scope > ${elementTag}`)).indexOf(item);
  
    const dragStart = (e: any) => {
      const item = getItemFromTarget(e.target as HTMLElement);
  
      item.classList.add(DRAG_CLASS, ...dragGhostClass);
  
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setData("source", getItemIndex(item).toString());
    };
  
    const dragOver = (e: any) => {
      e.preventDefault();
      // WARNING: This is called extremely often, so don't do any heavy lifting here
      e.dataTransfer.dropEffect = "move";
    };
  
    const dragEnter = (e: any) => {
      e.preventDefault();
  
      const draggedItem = parentEl!.querySelector(`.${DRAG_CLASS}`)!;
      const draggedOverItem = getItemFromTarget(e.target as HTMLElement);
      draggedItem.classList.remove(...dragGhostClass);
      draggedItem.classList.add(...dragClass);
      draggedOverItem.classList.add(DRAG_OVER_CLASS, ...dragOverClass);
    };
  
    const dragLeave = (e: any) => {
      getItemFromTarget(e.target as HTMLElement).classList.remove(DRAG_OVER_CLASS, ...dragOverClass);
    };
  
    const dragDrop = (e: any) => {
      e.preventDefault();
      e.stopPropagation();
  
      let dragged = getItemFromTarget(e.target as HTMLElement);
      let from = parseInt(e.dataTransfer.getData("source"));
      let to = getItemIndex(dragged);
  
      drag?.(from, to);
    };
  
    const dragEnd = () => {
      parentEl!.querySelectorAll(`.${DRAG_CLASS}`).forEach((item) => item.classList.remove(DRAG_CLASS, ...dragClass));
      parentEl!
        .querySelectorAll(`.${DRAG_OVER_CLASS}`)
        .forEach((item) => item.classList.remove(DRAG_OVER_CLASS, ...dragOverClass));
      if (handleClass) {
        parentEl!
          .querySelectorAll(':scope > [draggable="true"]')
          .forEach((item) => item.setAttribute("draggable", "false"));
      }
    };
  
    const handleMouseDown = (e: Event) => {
      const item = getItemFromTarget(e.target as HTMLElement);
      item.setAttribute("draggable", "true");
    };
  
    const handleMouseUp = (e: Event) => {
      const item = getItemFromTarget(e.target as HTMLElement);
      item.setAttribute("draggable", "false");
    };
  
    $effect(() => {
      if (itemEl) {
        if (handle) {
          handle.addEventListener("mousedown", handleMouseDown);
          handle.addEventListener("mouseup", handleMouseUp);
        } else if (!handleClass) {
          itemEl.setAttribute("draggable", "true");
        }
  
        return () => {
          handle?.removeEventListener("mousedown", handleMouseDown);
          handle?.removeEventListener("mouseup", handleMouseUp);
          itemEl?.removeAttribute("draggable");
        };
      }
    });
  </script>
  
  <svelte:element
    bind:this={itemEl}
    ondragend={dragEnd}
    ondragenter={dragEnter}
    ondragleave={dragLeave}
    ondragover={dragOver}
    ondragstart={dragStart}
    ondrop={dragDrop}
    role="row"
    tabindex="-1"
    this={elementTag}
    {...htmlAttributes}
  >
    {@render children()}
  </svelte:element>
  