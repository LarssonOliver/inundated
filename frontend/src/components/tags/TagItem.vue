<template>
  <div :class="containerClasses" @mouseenter="hover = true" @mouseleave="hover = false">
    <div>
      <span v-if="owner" class="owner-mark">{{ owner.prefix }}</span
      >{{ tag.name }}
    </div>
    <MaterialIcon
      @click="$emit('close', tag)"
      v-if="hover && canClose"
      class="close-icon"
      icon="close"
      size="1.2em"
    />
  </div>
</template>

<script setup lang="ts">
import MaterialIcon from "@/components/icons/MaterialIcon.vue";
import { shouldTextBeDarkFromBgColor } from "@/helpers/colors";
import { tagOwnerSpec } from "@/helpers/tagOwners";
import type { Tag } from "@/model";
import { computed, reactive, ref } from "vue";

defineEmits<{
  close: [tag: Tag];
}>();

const hover = ref(false);

const { tag, canClose } = defineProps<{
  tag: Tag;
  canClose?: boolean;
}>();

const darkText = computed(() => shouldTextBeDarkFromBgColor(tag.color));
// Must stay a computed, not a plain `tag.archived` read: reactive() only
// auto-unwraps refs on read, so a bare boolean captured here would freeze at
// mount time and stop tracking archive/unarchive changes on this instance
// (e.g. TagListView re-rendering the same TagItem after a toggle).
const isArchived = computed(() => tag.archived);
// Owned tags are drawn as an outlined pill with their owner kind's prefix
// ("#" for tasks), so they read as that kind of item wherever tags are shown.
const owner = computed(() => tagOwnerSpec(tag));
const containerClasses = reactive({
  "tag-container": true,
  "dark-text": computed(() => darkText.value && !owner.value),
  archived: isArchived,
  owned: computed(() => !!owner.value),
});
</script>

<style scoped>
.tag-container {
  background-color: v-bind("tag.color");
  margin: 0.25em 0.25em 0.25em 0;
  padding: 0.25em 0.6em;
  width: fit-content;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
}

.tag-container.owned {
  background-color: transparent;
  border: 2px solid v-bind("tag.color");
  padding: calc(0.25em - 2px) calc(0.6em - 2px);
}

.owner-mark {
  color: v-bind("tag.color");
  font-weight: 700;
  margin-right: 0.1em;
}

.tag-container.dark-text {
  color: var(--nord0);
}

.tag-container.archived div {
  text-decoration: line-through;
}

.close-icon {
  margin-left: 0.25em;
  cursor: pointer;
}
</style>
