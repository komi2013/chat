<script setup>
import { ref, computed } from "vue";

const props = defineProps({
  aliases: Array,
  groups: Array,
  modelValue: Array,
  placeholder: String
});
const emit = defineEmits(["update:modelValue"]);

const searchTerm = ref("");
const selectedAlias = ref([]);

if (Array.isArray(props.modelValue)) {
  selectedAlias.value = props.modelValue.map((name) => {
    const match = props.aliases.find((alias) => alias.aliasName === name);
    return match
      ? { id: match.aliasID, name: match.aliasName, image: match.aliasImg }
      : null;
  }).filter(Boolean);
}

// const emit = defineEmits(["update:selectedItems"]);

const filteredResults = computed(() => {
  if (!searchTerm.value || typeof searchTerm.value !== "string") return [];
  const aliasResults = props.aliases.map((item) => ({
    id: item.aliasID || "",
    name: item.aliasName || "",
    image: item.aliasImg || null,
  }));
  const groupResults = props.groups.map((group) => ({
    id: group.groupID || "",
    name: group.groupName || "",
    image: group.groupImg || null,
    aliasNames: group.aliasNames || []
  }));
  const results = [...aliasResults, ...groupResults].filter(
    (item) => item.name.includes(searchTerm.value)
  );
  return results.filter(
    (item) => !selectedAlias.value.some((selected) => selected.id === item.id)
  );
});

const addToSelection = (item) => {
  if (!selectedAlias.value.find((selected) => selected.id === item.id)) {
    if (item.aliasNames) {
      item.aliasNames.forEach((aliasName) => {
        const aliasMatch = props.aliases.find((alias) => alias.aliasName === aliasName);
        if (aliasMatch && !selectedAlias.value.some((selected) => selected.id === aliasMatch.aliasID)) {
          selectedAlias.value.push({
            id: aliasMatch.aliasID || "",
            name: aliasMatch.aliasName || "",
            image: aliasMatch.aliasImg || null,
          });
        }
      });
    } else {
      selectedAlias.value.push(item);
    }
    emitNames();
    searchTerm.value = "";
  }
};

const removeFromSelection = (item) => {
  selectedAlias.value = selectedAlias.value.filter(
    (selected) => selected.id !== item.id
  );
  emitNames();
};

const emitNames = (diff, item) => {
  const names = selectedAlias.value.map(item => item.name);
  emit("update:modelValue", names);
};


</script>

<template>
  <div>
    <input
      v-model="searchTerm"
      type="text"
      :placeholder="placeholder || 'ユーザー検索'"
    />
    <ul class="dropdown-menu">
      <li v-for="item in filteredResults" :key="item.id" @click="addToSelection(item)">

        <img v-if="item.image && item.image.charAt(0) != ','" 
          :src="item.image" class="min-icon">
        <span v-if="item.image && item.image.charAt(0) == ','"
          :style="'background-color:' + item.image.split(',')[2] "
          class="min-icon">
            <span>{{item.image.split(',')[1]}}</span>
        </span>

        <span>{{ item.name }}</span>
      </li>
    </ul>
    <br>
    <div v-for="item in selectedAlias" :key="item.id" class="selected-item">
      <img v-if="item.image && item.image.charAt(0) != ','" 
        :src="item.image" class="min-icon">
      <span v-if="item.image && item.image.charAt(0) == ','"
        class="min-icon"
        :style="'background-color:' + item.image.split(',')[2] ">
          <span>{{item.image.split(',')[1]}}</span>
      </span>
      <span>{{ item.name }}</span>
      <button @click="removeFromSelection(item)" class="remove-btn">x</button>
    </div>
  </div>
</template>

<style scoped>

.dropdown-menu {
  position: absolute;
  background-color: #fff;
  border: 1px solid #ccc;
  border-radius: 4px;
  z-index: 20;
  max-height: 200px;
  overflow-y: auto;
  list-style: none;
  padding: 0;
  margin: 0;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dropdown-menu li {
  padding: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  list-style-type: none;
  padding: 6px;
}

.dropdown-menu li:hover {
  background-color: #f0f0f0;
}


/*ul {
  list-style-type: none;
  padding: 0;
}

li {
  cursor: pointer;
  display: flex;
  align-items: center;
  margin-bottom: 5px;
}
*/
.min-icon {
  width: 26px;
  max-width: 26px;
  height: 26px;
  max-height: 26px;
  border-radius: 4px;
  display: inline-flex;
  vertical-align: middle;
  justify-content: center;
  align-items: center;
}

.selected-item {
  display: inline-block;
  border: 1px solid #3498db;
  margin: 5px 10px 5px 0;
  border-radius: 5px;
  padding: 3px;
}

/*img {
  margin-right: 10px;
  max-width: 20px;
  max-height: 20px;
}*/
</style>
