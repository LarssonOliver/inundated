<template>
  <div class="searchbox-container" v-if="!readOnly">
    <SelectDropdown
      placeholder=""
      searchable
      :creatable="canCreateFromCurrentQuery"
      manual-filter
      :options="tagSearchResult"
      :option-value="(tag) => tag.id"
      :option-label="(tag) => tag.name"
      @search="onTagSearch"
      @select="onTagSelect"
      @create="onTagCreate"
    >
      <template #default="{ option }">
        <div class="option">
          <TagItem :tag="option" />
          <span v-if="parentName(option)" class="option-parent">in {{ parentName(option) }}</span>
        </div>
      </template>
      <template #create="{ query }">
        <template v-if="taskNameFromQuery(query) !== null">
          Create task "{{ taskNameFromQuery(query) }}"...
        </template>
        <template v-else>Create "{{ query }}"...</template>
      </template>
    </SelectDropdown>
  </div>
  <div class="tag-list">
    <TagItem
      v-for="tag in tags"
      :key="tag.id"
      :tag="tag"
      :can-close="!readOnly"
      @close="onTagClose"
    />
  </div>
</template>

<script setup lang="ts">
import SelectDropdown from "@/components/inputs/SelectDropdown.vue";
import TagItem from "@/components/tags/TagItem.vue";
import type { Tag } from "@/model";
import { useTagsStore } from "@/stores/tags";
import { useTasksStore } from "@/stores/tasks";
import { computed, ref, watch } from "vue";

const model = defineModel<Set<string>>({ default: new Set<string>() });
const { readOnly, labelsOnly, allowTaskCreation } = defineProps<{
  readOnly?: boolean;
  /** Offer regular tags only, e.g. for a task's own tags. */
  labelsOnly?: boolean;
  /**
   * Lets typing "#name" search task tags only and create a new task from
   * this picker, e.g. for tagging a timespan to log time on a task. Off by
   * default so an unrelated picker (project tags, etc.) can't accidentally
   * create a task just because someone typed a leading "#".
   */
  allowTaskCreation?: boolean;
}>();

const tagsStore = useTagsStore();
const tasksStore = useTasksStore();
const tags = ref<Tag[]>([]);
// Raw, unfiltered server results for the current query. Filtered into
// tagSearchResult below so removing/adding a tag in `model` (e.g. via the
// pill's close button) re-excludes/re-includes it immediately, without
// waiting on the next keystroke to re-run the search.
const rawSearchResults = ref<Tag[]>([]);
const tagSearchResult = computed(() =>
  rawSearchResults.value
    .filter((tag) => !tag.archived && !model.value.has(tag.id))
    .filter((tag) => !labelsOnly || !tag.taskId)
    .slice(0, 8),
);

/**
 * Returns the task name typed after a leading "#", or null when the query
 * isn't a task query. A "#" prefix always narrows the search to task tags
 * (and is stripped from what's actually searched for) whenever task tags
 * aren't excluded outright by labelsOnly; creating a new task from an
 * unmatched "#query" additionally requires allowTaskCreation, guarded
 * separately below.
 */
function taskNameFromQuery(query: string): string | null {
  const trimmed = query.trim();
  if (labelsOnly || !trimmed.startsWith("#")) return null;
  return trimmed.slice(1).trim();
}

// Tracks the live search box text (updated per keystroke, ahead of the
// debounced search itself) so the "create" option can be hidden instantly
// for a "#query" when this picker isn't allowed to create tasks - showing
// "Create task ..." there, or falling back to literally creating a regular
// tag named "#query", would both be wrong.
const currentQuery = ref("");
const canCreateFromCurrentQuery = computed(
  () => taskNameFromQuery(currentQuery.value) === null || allowTaskCreation,
);

// Guards against overlapping searches: a slower response for an older query
// (including a search cleared out from under it) must not replace the
// results for the newer one.
let searchToken = 0;
// Loaded once so task options can show their parent's name; tracked
// separately from tasksStore.tasks.length, which stays 0 forever (and would
// otherwise re-trigger a fetch on every keystroke) when there really are no
// tasks yet. Reset on failure so a transient error doesn't disable the hint
// for the rest of this component's lifetime.
let tasksLoadAttempted = false;

async function search(query: string) {
  const token = ++searchToken;
  if (!query.trim()) {
    rawSearchResults.value = [];
    return;
  }

  const taskName = taskNameFromQuery(query);
  const kind = labelsOnly ? "label" : taskName !== null ? "task" : "all";
  const results = await tagsStore.searchTagsOnServer(taskName ?? query, kind);
  if (kind !== "label" && !tasksLoadAttempted) {
    tasksLoadAttempted = true;
    tasksStore.fetchTasks().catch(() => {
      tasksLoadAttempted = false;
    });
  }
  if (token !== searchToken) return;

  rawSearchResults.value = results;
}

const SEARCH_DEBOUNCE_MS = 200;
let searchDebounceTimer: ReturnType<typeof setTimeout> | undefined;

function debouncedSearch(query: string) {
  clearTimeout(searchDebounceTimer);
  searchDebounceTimer = setTimeout(() => void search(query), SEARCH_DEBOUNCE_MS);
}

function parentName(tag: Tag): string | undefined {
  if (!tag.taskId) return undefined;
  const parentId = tasksStore.getTaskById(tag.taskId)?.parentId;
  return parentId ? tasksStore.getTaskById(parentId)?.name : undefined;
}

// Guards against overlapping refreshes: if the model changes again before an
// earlier refresh's (slower) network calls resolve, that stale result must
// not clobber the tags the newer refresh already resolved.
let refreshToken = 0;

watch(model, async () => await refreshTags(), { deep: true, immediate: true });

async function refreshTags() {
  const token = ++refreshToken;
  await tagsStore.fetchTags();
  const resolved = await Promise.all(
    [...model.value].map((id) => tagsStore.getTagById(id) ?? fetchAssignedTag(id)),
  );
  if (token !== refreshToken) return; // superseded by a newer refresh
  tags.value = resolved.filter((tag): tag is Tag => tag != null);
}

// Tags already assigned to this item must still be shown even if archived,
// but the shared tags cache only holds non-archived tags unless the "show
// archived" toggle is on elsewhere, so fall back to fetching them directly.
// Task tags are never in that cache, so they're always fetched this way.
// Uses the non-detailed fetch since this pill display has no use for stats
// like totalTimeMs, which would otherwise make the server aggregate them
// for nothing.
async function fetchAssignedTag(id: string): Promise<Tag | undefined> {
  try {
    return await tagsStore.fetchTagById(id);
  } catch {
    return undefined;
  }
}

function onTagSearch(query: string) {
  currentQuery.value = query;
  if (!query.trim()) {
    clearTimeout(searchDebounceTimer);
    searchToken++; // invalidate any in-flight search so it can't repopulate results
    rawSearchResults.value = [];
    return;
  }
  debouncedSearch(query);
}

function onTagSelect(tag: Tag) {
  model.value = new Set([...model.value, tag.id]);
}

function onTagClose(tag: Tag) {
  model.value = new Set([...model.value].filter((id) => id !== tag.id));
}

async function onTagCreate(query: string) {
  const taskName = taskNameFromQuery(query);
  if (taskName !== null) {
    // The create row is hidden in this case (see canCreateFromCurrentQuery),
    // so this only guards against it somehow still firing.
    if (!allowTaskCreation || !taskName) return;
    const task = await tasksStore.createTaskFromName(taskName);
    // Fetch the new task tag so its pill can show before the next refresh.
    await tagsStore.fetchTagById(task.tagId);
    model.value = new Set([...model.value, task.tagId]);
    return;
  }

  const name = query.trim();
  if (!name) return;
  const tag = await tagsStore.createTagFromName(name);
  if (tag) {
    model.value = new Set([...model.value, tag.id]);
  }
}
</script>

<style scoped>
.tag-list {
  display: flex;
  flex-wrap: wrap;
  margin-right: 0.5em;
}

.option {
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.option-parent {
  color: var(--nord3);
  font-size: 0.85em;
}

.searchbox-container {
  display: flex;
  align-items: center;
  min-width: 8em;
  margin-left: 0.5em;
  margin-right: 0.5em;
  --max-width: 12em;
}
</style>
