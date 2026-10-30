/**
 * Zenith TaskFlow - Client Application Logic
 * Integrates with Go 1.27 backend REST API
 */

(() => {
  'use strict';

  // Application State
  const state = {
    todos: [],
    stats: {
      total: 0,
      active: 0,
      completed: 0,
      overdue: 0,
      high_priority: 0,
      medium_priority: 0,
      low_priority: 0,
    },
    filter: 'all',
    priority: 'all',
    category: 'all',
    sort: 'created_desc',
    search: '',
    pendingDeleteId: null,
  };

  // DOM Elements
  const elements = {
    taskList: document.getElementById('task-list'),
    loadingSpinner: document.getElementById('loading-spinner'),
    emptyState: document.getElementById('empty-state'),
    searchInput: document.getElementById('search-input'),
    themeToggleBtn: document.getElementById('theme-toggle-btn'),
    openCreateModalBtn: document.getElementById('open-create-modal-btn'),
    emptyCreateBtn: document.getElementById('empty-create-btn'),
    clearCompletedBtn: document.getElementById('clear-completed-btn'),
    quickAddForm: document.getElementById('quick-add-form'),
    quickTaskTitle: document.getElementById('quick-task-title'),
    quickTaskCategory: document.getElementById('quick-task-category'),
    quickTaskPriority: document.getElementById('quick-task-priority'),
    quickTaskDate: document.getElementById('quick-task-date'),
    filterTabs: document.querySelectorAll('.filter-tab'),
    filterPriority: document.getElementById('filter-priority'),
    filterCategory: document.getElementById('filter-category'),
    sortSelect: document.getElementById('sort-select'),
    // Stats Elements
    statTotal: document.getElementById('stat-total'),
    statActive: document.getElementById('stat-active'),
    statCompleted: document.getElementById('stat-completed'),
    statRate: document.getElementById('stat-rate'),
    statProgressFill: document.getElementById('stat-progress-fill'),
    badgeAll: document.getElementById('badge-all'),
    badgeActive: document.getElementById('badge-active'),
    badgeCompleted: document.getElementById('badge-completed'),
    footerCount: document.getElementById('footer-count'),
    footerDb: document.getElementById('footer-db'),
    // Modals
    taskModal: document.getElementById('task-modal'),
    taskForm: document.getElementById('task-form'),
    modalHeading: document.getElementById('modal-heading'),
    modalSubheading: document.getElementById('modal-subheading'),
    taskIdInput: document.getElementById('task-id'),
    taskTitleInput: document.getElementById('task-title'),
    taskDescInput: document.getElementById('task-desc'),
    taskCategorySelect: document.getElementById('task-category'),
    taskPrioritySelect: document.getElementById('task-priority'),
    taskDueDateInput: document.getElementById('task-due-date'),
    statusToggleGroup: document.getElementById('status-toggle-group'),
    taskCompletedToggle: document.getElementById('task-completed-toggle'),
    closeModalBtn: document.getElementById('close-modal-btn'),
    cancelModalBtn: document.getElementById('cancel-modal-btn'),
    saveBtnText: document.getElementById('save-btn-text'),
    // Delete Modal
    deleteModal: document.getElementById('delete-modal'),
    deleteTaskTitleSpan: document.getElementById('delete-task-title'),
    cancelDeleteBtn: document.getElementById('cancel-delete-btn'),
    confirmDeleteBtn: document.getElementById('confirm-delete-btn'),
    toastContainer: document.getElementById('toast-container'),
  };

  /* ==========================================================================
     API Client
     ========================================================================== */

  async function apiRequest(endpoint, options = {}) {
    try {
      const res = await fetch(endpoint, {
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        ...options,
      });

      const data = await res.json();
      if (!res.ok || !data.success) {
        throw new Error(data.message || `Request failed with status ${res.status}`);
      }
      return data;
    } catch (err) {
      console.error(`API Error [${endpoint}]:`, err);
      showToast(err.message || 'Network request failed', 'error');
      throw err;
    }
  }

  async function loadData() {
    elements.loadingSpinner.classList.remove('hidden');
    try {
      const query = new URLSearchParams({
        q: state.search,
        status: state.filter,
        priority: state.priority,
        category: state.category,
        sort: state.sort,
      });

      const [todosRes, statsRes] = await Promise.all([
        apiRequest(`/api/todos?${query.toString()}`),
        apiRequest('/api/stats'),
      ]);

      state.todos = todosRes.data || [];
      state.stats = statsRes.data || state.stats;

      render();
    } catch (err) {
      console.error('Failed to load tasks:', err);
    } finally {
      elements.loadingSpinner.classList.add('hidden');
    }
  }

  /* ==========================================================================
     Render Functions
     ========================================================================== */

  function render() {
    renderStats();
    renderTasks();
  }

  function renderStats() {
    const { total, active, completed } = state.stats;
    const rate = total > 0 ? Math.round((completed / total) * 100) : 0;

    elements.statTotal.textContent = total;
    elements.statActive.textContent = active;
    elements.statCompleted.textContent = completed;
    elements.statRate.textContent = `${rate}%`;
    elements.statProgressFill.style.width = `${rate}%`;

    elements.badgeAll.textContent = total;
    elements.badgeActive.textContent = active;
    elements.badgeCompleted.textContent = completed;

    elements.footerCount.textContent = `${state.todos.length} item${state.todos.length === 1 ? '' : 's'} displayed`;

    if (elements.footerDb && state.stats.database_type) {
      elements.footerDb.textContent = state.stats.database_type === 'PostgreSQL'
        ? '🐘 PostgreSQL Active'
        : `📁 ${state.stats.database_type}`;
    }
  }

  function renderTasks() {
    elements.taskList.innerHTML = '';

    if (state.todos.length === 0) {
      elements.emptyState.classList.remove('hidden');
      elements.taskList.classList.add('hidden');
      return;
    }

    elements.emptyState.classList.add('hidden');
    elements.taskList.classList.remove('hidden');

    const fragment = document.createDocumentFragment();

    state.todos.forEach(todo => {
      const li = document.createElement('li');
      li.className = `task-item priority-${todo.priority} ${todo.completed ? 'completed' : ''}`;
      li.dataset.id = todo.id;

      // Format Due Date and status
      const dueDateBadge = getDueDateBadge(todo.due_date, todo.completed);
      const categoryClass = getCategoryClass(todo.category);

      li.innerHTML = `
        <div class="checkbox-wrap">
          <input type="checkbox" 
                 class="custom-checkbox" 
                 ${todo.completed ? 'checked' : ''} 
                 aria-label="Mark task '${escapeHtml(todo.title)}' as ${todo.completed ? 'incomplete' : 'complete'}">
        </div>
        <div class="task-body">
          <div class="task-header-row">
            <h3 class="task-title" title="Click to edit">${escapeHtml(todo.title)}</h3>
          </div>
          ${todo.description ? `<p class="task-desc">${escapeHtml(todo.description)}</p>` : ''}
          <div class="task-meta-row">
            <span class="meta-chip chip-category ${categoryClass}">
              ${escapeHtml(todo.category)}
            </span>
            <span class="meta-chip chip-priority priority-chip-${todo.priority}">
              ${todo.priority}
            </span>
            ${dueDateBadge}
          </div>
        </div>
        <div class="task-actions">
          <button class="action-btn edit-btn" title="Edit task" aria-label="Edit task">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
            </svg>
          </button>
          <button class="action-btn delete-btn" title="Delete task" aria-label="Delete task">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
          </button>
        </div>
      `;

      // Event Listeners for Item
      const checkbox = li.querySelector('.custom-checkbox');
      checkbox.addEventListener('change', () => toggleTaskCompletion(todo.id, li));

      const titleEl = li.querySelector('.task-title');
      titleEl.addEventListener('click', () => openEditModal(todo));

      const editBtn = li.querySelector('.edit-btn');
      editBtn.addEventListener('click', () => openEditModal(todo));

      const deleteBtn = li.querySelector('.delete-btn');
      deleteBtn.addEventListener('click', () => promptDelete(todo));

      fragment.appendChild(li);
    });

    elements.taskList.appendChild(fragment);
  }

  function getDueDateBadge(dueDateStr, completed) {
    if (!dueDateStr) return '';

    const todayStr = new Date().toISOString().split('T')[0];
    let badgeClass = '';
    let label = dueDateStr;

    if (!completed) {
      if (dueDateStr < todayStr) {
        badgeClass = 'is-overdue';
        label = `⚠️ Overdue (${dueDateStr})`;
      } else if (dueDateStr === todayStr) {
        badgeClass = 'is-today';
        label = `📅 Today`;
      } else {
        label = `📅 ${dueDateStr}`;
      }
    } else {
      label = `📅 ${dueDateStr}`;
    }

    return `<span class="meta-chip chip-date ${badgeClass}">${escapeHtml(label)}</span>`;
  }

  function getCategoryClass(cat) {
    const map = {
      'work': 'category-work',
      'backend': 'category-backend',
      'frontend': 'category-frontend',
      'design': 'category-design',
      'health': 'category-health',
      'study': 'category-study',
    };
    return map[String(cat).toLowerCase()] || 'chip-category';
  }

  function escapeHtml(str) {
    if (!str) return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  /* ==========================================================================
     Task Actions
     ========================================================================== */

  async function toggleTaskCompletion(id, elementItem) {
    try {
      // Optimistic update
      elementItem.classList.toggle('completed');

      const res = await apiRequest(`/api/todos/${id}/toggle`, {
        method: 'PATCH',
      });

      showToast(
        res.data.completed ? '🎉 Task completed!' : 'Task marked as active',
        'info'
      );
      loadData();
    } catch (err) {
      // Revert if error
      elementItem.classList.toggle('completed');
    }
  }

  async function handleQuickAdd(e) {
    e.preventDefault();
    const title = elements.quickTaskTitle.value.trim();
    if (!title) return;

    const payload = {
      title,
      category: elements.quickTaskCategory.value,
      priority: elements.quickTaskPriority.value,
      due_date: elements.quickTaskDate.value || '',
      description: '',
    };

    try {
      await apiRequest('/api/todos', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      elements.quickTaskTitle.value = '';
      elements.quickTaskDate.value = '';
      showToast('⚡ Task added in a flash!', 'success');
      loadData();
    } catch (err) {
      // Toast handled in apiRequest
    }
  }

  function openCreateModal() {
    elements.taskForm.reset();
    elements.taskIdInput.value = '';
    elements.modalHeading.textContent = 'Create New Task';
    elements.modalSubheading.textContent = 'Capture ideas and schedule deliverables';
    elements.saveBtnText.textContent = 'Create Task';
    elements.statusToggleGroup.classList.add('hidden');
    elements.taskPrioritySelect.value = 'medium';
    elements.taskCategorySelect.value = 'General';

    // Set default due date to today
    elements.taskDueDateInput.value = new Date().toISOString().split('T')[0];

    elements.taskModal.showModal();
    elements.taskTitleInput.focus();
  }

  function openEditModal(todo) {
    elements.taskIdInput.value = todo.id;
    elements.taskTitleInput.value = todo.title;
    elements.taskDescInput.value = todo.description || '';
    elements.taskCategorySelect.value = todo.category || 'General';
    elements.taskPrioritySelect.value = todo.priority || 'medium';
    elements.taskDueDateInput.value = todo.due_date || '';
    elements.taskCompletedToggle.checked = !!todo.completed;

    elements.modalHeading.textContent = 'Edit Task';
    elements.modalSubheading.textContent = `Modifying task ID: ${todo.id}`;
    elements.saveBtnText.textContent = 'Save Changes';
    elements.statusToggleGroup.classList.remove('hidden');

    elements.taskModal.showModal();
    elements.taskTitleInput.focus();
  }

  async function handleSaveTask(e) {
    e.preventDefault();
    const id = elements.taskIdInput.value;
    const title = elements.taskTitleInput.value.trim();

    if (!title) {
      showToast('Task title is required', 'error');
      elements.taskTitleInput.focus();
      return;
    }

    const payload = {
      title,
      description: elements.taskDescInput.value.trim(),
      category: elements.taskCategorySelect.value,
      priority: elements.taskPrioritySelect.value,
      due_date: elements.taskDueDateInput.value || '',
    };

    try {
      if (id) {
        // Edit mode
        payload.completed = elements.taskCompletedToggle.checked;
        await apiRequest(`/api/todos/${id}`, {
          method: 'PUT',
          body: JSON.stringify(payload),
        });
        showToast('Task updated successfully', 'success');
      } else {
        // Create mode
        await apiRequest('/api/todos', {
          method: 'POST',
          body: JSON.stringify(payload),
        });
        showToast('New task created!', 'success');
      }

      elements.taskModal.close();
      loadData();
    } catch (err) {
      // Toast handled
    }
  }

  function promptDelete(todo) {
    state.pendingDeleteId = todo.id;
    elements.deleteTaskTitleSpan.textContent = `"${todo.title}"`;
    elements.deleteModal.showModal();
    elements.confirmDeleteBtn.focus();
  }

  async function confirmDelete() {
    if (!state.pendingDeleteId) return;

    try {
      await apiRequest(`/api/todos/${state.pendingDeleteId}`, {
        method: 'DELETE',
      });
      showToast('Task deleted', 'info');
      elements.deleteModal.close();
      state.pendingDeleteId = null;
      loadData();
    } catch (err) {
      // Toast handled
    }
  }

  async function handleClearCompleted() {
    if (state.stats.completed === 0) {
      showToast('No completed tasks to clear', 'info');
      return;
    }

    try {
      const res = await apiRequest('/api/todos/clear-completed', {
        method: 'POST',
      });
      const count = res.data?.cleared_count || 0;
      showToast(`Cleared ${count} completed task${count === 1 ? '' : 's'}!`, 'success');
      loadData();
    } catch (err) {
      // Toast handled
    }
  }

  /* ==========================================================================
     Toast Notifications
     ========================================================================== */

  function showToast(message, type = 'info', duration = 3200) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;

    const iconMap = {
      success: '✓',
      error: '✕',
      info: '⚡',
    };

    toast.innerHTML = `
      <span class="toast-icon">${iconMap[type] || '•'}</span>
      <span class="toast-msg">${escapeHtml(message)}</span>
    `;

    elements.toastContainer.appendChild(toast);

    setTimeout(() => {
      toast.classList.add('toast-exit');
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
      }, 300);
    }, duration);
  }

  /* ==========================================================================
     Theme Switcher
     ========================================================================== */

  function initTheme() {
    const savedTheme = localStorage.getItem('taskflow_theme') || 'dark';
    applyTheme(savedTheme);

    elements.themeToggleBtn.addEventListener('click', () => {
      const current = document.body.classList.contains('theme-light') ? 'light' : 'dark';
      const next = current === 'dark' ? 'light' : 'dark';
      applyTheme(next);
      localStorage.setItem('taskflow_theme', next);
    });
  }

  function applyTheme(theme) {
    if (theme === 'light') {
      document.body.classList.remove('theme-dark');
      document.body.classList.add('theme-light');
    } else {
      document.body.classList.remove('theme-light');
      document.body.classList.add('theme-dark');
    }
  }

  /* ==========================================================================
     Modal Fallback & Event Handlers
     ========================================================================== */

  function setupDialogPolyfillAndEvents() {
    // Light dismiss fallback per modern-web-guidance
    if (!('closedBy' in HTMLDialogElement.prototype)) {
      const outside = (e, dialog) => {
        const r = dialog.getBoundingClientRect();
        return e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom;
      };
      for (const dialog of document.querySelectorAll('dialog[closedby="any"]')) {
        let pressedOutside = false;
        dialog.addEventListener('pointerdown', (e) => { pressedOutside = outside(e, dialog); });
        dialog.addEventListener('pointerup', (e) => { if (pressedOutside && outside(e, dialog)) dialog.close(); });
      }
    }

    elements.openCreateModalBtn.addEventListener('click', openCreateModal);
    elements.emptyCreateBtn.addEventListener('click', openCreateModal);
    elements.closeModalBtn.addEventListener('click', () => elements.taskModal.close());
    elements.cancelModalBtn.addEventListener('click', () => elements.taskModal.close());
    elements.taskForm.addEventListener('submit', handleSaveTask);

    elements.cancelDeleteBtn.addEventListener('click', () => {
      elements.deleteModal.close();
      state.pendingDeleteId = null;
    });
    elements.confirmDeleteBtn.addEventListener('click', confirmDelete);

    elements.clearCompletedBtn.addEventListener('click', handleClearCompleted);
    elements.quickAddForm.addEventListener('submit', handleQuickAdd);

    // Filter tabs
    elements.filterTabs.forEach(tab => {
      tab.addEventListener('click', () => {
        elements.filterTabs.forEach(t => {
          t.classList.remove('active');
          t.setAttribute('aria-selected', 'false');
        });
        tab.classList.add('active');
        tab.setAttribute('aria-selected', 'true');
        state.filter = tab.dataset.filter;
        loadData();
      });
    });

    // Select filters & sort
    elements.filterPriority.addEventListener('change', (e) => {
      state.priority = e.target.value;
      loadData();
    });

    elements.filterCategory.addEventListener('change', (e) => {
      state.category = e.target.value;
      loadData();
    });

    elements.sortSelect.addEventListener('change', (e) => {
      state.sort = e.target.value;
      loadData();
    });

    // Search debounce
    let searchDebounceTimeout = null;
    elements.searchInput.addEventListener('input', (e) => {
      clearTimeout(searchDebounceTimeout);
      searchDebounceTimeout = setTimeout(() => {
        state.search = e.target.value;
        loadData();
      }, 250);
    });

    // Keyboard Shortcuts: '/' for search, 'N' for new task, 'Escape' to dismiss
    window.addEventListener('keydown', (e) => {
      const activeTag = document.activeElement ? document.activeElement.tagName.toLowerCase() : '';
      const isInput = activeTag === 'input' || activeTag === 'textarea' || activeTag === 'select';

      if (e.key === '/' && !isInput) {
        e.preventDefault();
        elements.searchInput.focus();
        elements.searchInput.select();
      } else if ((e.key === 'n' || e.key === 'N') && !isInput) {
        e.preventDefault();
        openCreateModal();
      }
    });
  }

  /* ==========================================================================
     Initialization
     ========================================================================== */

  document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    setupDialogPolyfillAndEvents();
    loadData();
  });
})();
