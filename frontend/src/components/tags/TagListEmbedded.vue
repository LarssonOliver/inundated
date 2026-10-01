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
        <template v-if="ownerQuery(query)">
          Create {{ ownerQuery(query)?.spec.noun }} "{{ ownerQuery(query)?.name }}"...
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
import { allTagOwnerKinds, isOwnedBy, parseOwnerQuery } from "@/helpers/tagOwners";
import type { TagKind } from "@/api";
import type { Tag, TagOwnerKind } from "@/model";
import { useTagsStore } from "@/stores/tags";
import { useTasksStore } from "@/stores/tasks";
import { computed, ref, watch } from "vue";

const model = defineModel<Set<string>>({ default: new Set<string>() });
const {
  readOnly,
  ownerKinds = allTagOwnerKinds,
  createOwners = [],
  hideOwnerKinds = [],
} = defineProps<{
  readOnly?: boolean;
  /**
   * The kinds of owned tags offered besides regular tags, matching what the
   * item being tagged may carry: e.g. a task takes project tags but not
   * task tags, and a project takes task tags but not project tags. Every
   * kind by default; none for regular tags only.
   */
  ownerKinds?: readonly TagOwnerKind[];
  /**
   * The kinds of owners this picker may create from "prefix + name" (e.g.
   * "#name" for a task, "@name" for a project), for tagging a timespan to
   * log time on a new task. None by default, so an unrelated picker can't
   * accidentally create one just because someone typed a leading prefix.
   */
  createOwners?: readonly TagOwnerKind[];
  /**
   * The kinds of owned tags left out of this list's pills, for a page that
   * shows them elsewhere: e.g. the project page lists its task tags as its
   * task list. They stay in the model, so editing the visible tags keeps
   * them, and the picker still offers and creates them as ownerKinds and
   * createOwners allow. Only hide a kind the page shows elsewhere, since a
   * tag of it picked here would otherwise seem to vanish. None by default.
   */
  hideOwnerKinds?: readonly TagOwnerKind[];
}>();

function isHidden(tag: Tag): boolean {
  return !!tag.owner && hideOwnerKinds.includes(tag.owner.kind);
}

const tagsStore = useTagsStore();
const tasksStore = useTasksStore();
// The assigned tags as last resolved by refreshTags. Pills show the store's
// current copy of each instead, so a refetch elsewhere (e.g. a task tag's
// color after its task's tags change, or a renamed project) reaches pills
// already on screen. The resolved copy is only a fallback for a tag the
// store has since dropped from its listing (e.g. an archived tag after the
// archived filter changes), and a tag the store knows was deleted (e.g.
// with its project) isn't shown at all.
const resolvedTags = ref<Tag[]>([]);
const tags = computed(() =>
  resolvedTags.value
    .filter((tag) => !tagsStore.isTagDeleted(tag) && !isHidden(tag))
    .map((tag) => tagsStore.getTagById(tag.id) ?? tag),
);
// Raw, unfiltered server results for the current query. Filtered into
// tagSearchResult below so removing/adding a tag in `model` (e.g. via the
// pill's close button) re-excludes/re-includes it immediately, without
// waiting on the next keystroke to re-run the search.
const rawSearchResults = ref<Tag[]>([]);
const tagSearchResult = computed(() =>
  rawSearchResults.value
    .filter((tag) => !tag.archived && !model.value.has(tag.id))
    .filter((tag) => !tag.owner || ownerKinds.includes(tag.owner.kind))
    .slice(0, 8),
);

/**
 * Returns the owner kind and name typed after a leading owner prefix ("#"
 * for tasks, "@" for projects, see helpers/tagOwners), or null when the
 * query has none of a kind this picker offers. Such a prefix narrows the
 * search to that kind's tags (and is stripped from what's actually
 * searched for). Creating a new owner from an unmatched query additionally
 * requires createOwners, guarded separately below.
 */
function ownerQuery(query: string) {
  return parseOwnerQuery(query, ownerKinds);
}

/**
 * Whether the query starts with the prefix of an owner kind this picker
 * doesn't offer. It then only finds regular tags named with the prefix,
 * and creates nothing: a regular tag named e.g. "@name" would look like a
 * project tag.
 */
function namesUnofferedOwner(query: string): boolean {
  return !ownerQuery(query) && parseOwnerQuery(query) !== null;
}

/** Whether this picker may create an owner of the kind the query names. */
function canCreateOwner(owner: NonNullable<ReturnType<typeof ownerQuery>>): boolean {
  return createOwners.includes(owner.spec.kind);
}

// Tracks the live search box text (updated per keystroke, ahead of the
// debounced search itself) so the "create" option can be hidden instantly
// for a "#query" when this picker isn't allowed to create tasks - showing
// "Create task ..." there, or falling back to literally creating a regular
// tag named "#query", would both be wrong. Likewise for "@query" and
// projects, and for the prefix of a kind this picker doesn't offer at all
// (see namesUnofferedOwner).
const currentQuery = ref("");
const canCreateFromCurrentQuery = computed(() => {
  const owner = ownerQuery(currentQuery.value);
  if (owner) return canCreateOwner(owner);
  return !namesUnofferedOwner(currentQuery.value);
});

// The kinds of tags a search without an owner prefix asks the server for:
// regular tags and each owned kind offered, which take turns filling the
// server's results so one kind's matches can't crowd another's out. Kinds
// the picker doesn't offer aren't searched at all.
const searchedKinds = computed<TagKind[]>(() => ["label", ...ownerKinds]);

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
  if (!query) {
    rawSearchResults.value = [];
    return;
  }

  const owner = ownerQuery(query);
  let kinds: TagKind[] = searchedKinds.value;
  if (owner) kinds = [owner.spec.kind];
  else if (namesUnofferedOwner(query)) kinds = ["label"];
  const results = await tagsStore.searchTagsOnServer(owner?.name ?? query, kinds);
  const mayShowTaskTags = kinds.includes("task");
  if (mayShowTaskTags && !tasksLoadAttempted) {
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
  if (!tag.owner || !isOwnedBy(tag, "task")) return undefined;
  const parentId = tasksStore.getTaskById(tag.owner.id)?.parentId;
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
  resolvedTags.value = resolved.filter((tag): tag is Tag => tag != null);
}

// Tags already assigned to this item must still be shown even if archived,
// but the shared tags cache only holds non-archived tags unless the "show
// archived" toggle is on elsewhere, so fall back to fetching them directly.
// Owned tags are never in that cache, so they're always fetched this way.
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
  if (!query) {
    clearTimeout(searchDebounceTimer);
    searchToken++; // invalidate any in-flight search so it can't repopulate results
    rawSearchResults.value = [];
    return;
  }
  // A whitespace-only query still reaches search() -> searchTagsOnServer,
  // which trims it to "" and treats that as "browse all" - only a literally
  // empty box should skip the request and clear results outright.
  debouncedSearch(query);
}

function onTagSelect(tag: Tag) {
  model.value = new Set([...model.value, tag.id]);
}

function onTagClose(tag: Tag) {
  model.value = new Set([...model.value].filter((id) => id !== tag.id));
}

async function onTagCreate(query: string) {
  const owner = ownerQuery(query);
  if (owner) {
    // The create row is hidden when this picker can't create this kind
    // (see canCreateFromCurrentQuery), so this only guards against it
    // somehow still firing.
    if (!canCreateOwner(owner) || !owner.name) return;
    const tagId = await owner.spec.create(owner.name);
    // Fetch the new owned tag so its pill can show before the next refresh.
    await tagsStore.fetchTagById(tagId);
    model.value = new Set([...model.value, tagId]);
    return;
  }

  const name = query.trim();
  if (!name || namesUnofferedOwner(name)) return;
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
