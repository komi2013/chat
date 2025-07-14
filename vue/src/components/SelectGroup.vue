<template>
  <div class="dropdown-menu">
    <div v-if="groups.length === 0" class="dropdown-item">
      グループはありません
    </div>
    <div 
      v-for="group in groups"
      :key="group.groupID"
      class="dropdown-item" 
      :class="{ 'selected': selectedGroup && group.groupID === selectedGroup.groupID }"
      @click="selectGroup(group)"
    >
      <img v-if="group.groupImg && group.groupImg.charAt(0) != ','" 
        :src="group.groupImg" class="min-icon">
      <span v-if="group.groupImg && group.groupImg.charAt(0) == ','"
        :style="'background-color:' + group.groupImg.split(',')[2] "
        class="min-icon">
          <span>{{group.groupImg.split(',')[1]}}</span>
      </span>
      <span>{{ group.groupName }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  groups: {
    type: Array,
    required: true,
  },
  modelValue: {
    type: Array,
    required: true,
  },
});

const emit = defineEmits(['update:modelValue']);

const selectedGroup = ref(props.modelValue);

function selectGroup(group) {
  selectedGroup.value = group;
  emit('update:modelValue', group);
}
</script>

<style scoped>
.dropdown-menu {
  border: 1px solid #ccc;
  border-radius: 4px;
  width: 200px;
  padding: 5px;
  background-color: #fff;
}

.dropdown-item {
  padding: 8px;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.dropdown-item.selected {
  background-color: #e6f7ff;
}

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


</style>
