// E-Voting System JavaScript

document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });

    // Initialize popovers
    var popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    var popoverList = popoverTriggerList.map(function (popoverTriggerEl) {
        return new bootstrap.Popover(popoverTriggerEl);
    });

    // Auto-hide alerts after 5 seconds
    setTimeout(function() {
        var alerts = document.querySelectorAll('.alert:not(.alert-permanent)');
        alerts.forEach(function(alert) {
            var bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        });
    }, 5000);

    // Form validation
    var forms = document.querySelectorAll('.needs-validation');
    Array.prototype.slice.call(forms).forEach(function(form) {
        form.addEventListener('submit', function(event) {
            if (!form.checkValidity()) {
                event.preventDefault();
                event.stopPropagation();
            }
            form.classList.add('was-validated');
        }, false);
    });

    // Candidate selection for voting
    initializeCandidateSelection();
    
    // Token management
    initializeTokenManagement();
    
    // Date/time validation
    initializeDateTimeValidation();
    
    // Confirmation dialogs
    initializeConfirmationDialogs();
});

// Candidate Selection Functions
function initializeCandidateSelection() {
    const candidateRadios = document.querySelectorAll('input[name="candidate_id"]');
    
    candidateRadios.forEach(function(radio) {
        radio.addEventListener('change', function() {
            // Remove selection styling from all cards
            document.querySelectorAll('.candidate-card').forEach(function(card) {
                card.classList.remove('border-primary', 'bg-light', 'selected');
            });
            
            // Add selection styling to selected card
            if (this.checked) {
                const card = this.closest('.candidate-card');
                if (card) {
                    card.classList.add('border-primary', 'bg-light', 'selected');
                }
            }
        });
        
        // Make entire card clickable
        const card = radio.closest('.candidate-card');
        if (card) {
            card.addEventListener('click', function(e) {
                if (e.target.type !== 'radio') {
                    radio.checked = true;
                    radio.dispatchEvent(new Event('change'));
                }
            });
        }
    });
}

// Token Management Functions
function initializeTokenManagement() {
    // Copy token functionality
    window.copyToken = function(token) {
        if (navigator.clipboard) {
            navigator.clipboard.writeText(token).then(function() {
                showCopySuccess();
            }).catch(function() {
                fallbackCopyToken(token);
            });
        } else {
            fallbackCopyToken(token);
        }
    };
    
    // Copy all tokens
    window.copyAllTokens = function() {
        const tokens = Array.from(document.querySelectorAll('.token-code')).map(el => el.textContent);
        const tokenText = tokens.join('\n');
        
        if (navigator.clipboard) {
            navigator.clipboard.writeText(tokenText).then(function() {
                showNotification('All tokens copied to clipboard!', 'success');
            });
        } else {
            fallbackCopyToken(tokenText);
        }
    };
    
    // Export tokens as CSV
    window.exportTokens = function() {
        const tokens = Array.from(document.querySelectorAll('tbody tr')).map(row => {
            const cells = row.querySelectorAll('td');
            return {
                token: cells[0].querySelector('.token-code').textContent,
                status: cells[1].textContent.trim(),
                created: cells[2].textContent.trim(),
                used: cells[3].textContent.trim()
            };
        });
        
        const csv = 'Token,Status,Created,Used\n' + 
                   tokens.map(t => `"${t.token}","${t.status}","${t.created}","${t.used}"`).join('\n');
        
        downloadCSV(csv, 'voting-tokens.csv');
    };
}

// Date/Time Validation
function initializeDateTimeValidation() {
    const startDateInput = document.getElementById('start_date');
    const endDateInput = document.getElementById('end_date');
    
    if (startDateInput && endDateInput) {
        function validateDates() {
            const startDate = new Date(startDateInput.value);
            const endDate = new Date(endDateInput.value);
            const now = new Date();
            
            // Clear previous validation
            startDateInput.setCustomValidity('');
            endDateInput.setCustomValidity('');
            
            // Validate start date is not in the past
            if (startDate < now) {
                startDateInput.setCustomValidity('Start date cannot be in the past');
            }
            
            // Validate end date is after start date
            if (endDate <= startDate) {
                endDateInput.setCustomValidity('End date must be after start date');
            }
        }
        
        startDateInput.addEventListener('change', validateDates);
        endDateInput.addEventListener('change', validateDates);
    }
}

// Confirmation Dialogs
function initializeConfirmationDialogs() {
    // Delete confirmations
    document.querySelectorAll('[data-confirm]').forEach(function(element) {
        element.addEventListener('click', function(e) {
            const message = this.getAttribute('data-confirm');
            if (!confirm(message)) {
                e.preventDefault();
                return false;
            }
        });
    });
}

// Utility Functions
function showCopySuccess() {
    showNotification('Token copied to clipboard!', 'success');
}

function fallbackCopyToken(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.select();
    
    try {
        document.execCommand('copy');
        showNotification('Token copied to clipboard!', 'success');
    } catch (err) {
        showNotification('Failed to copy token', 'error');
    }
    
    document.body.removeChild(textArea);
}

function showNotification(message, type = 'info') {
    const alertClass = type === 'success' ? 'alert-success' : 
                      type === 'error' ? 'alert-danger' : 
                      type === 'warning' ? 'alert-warning' : 'alert-info';
    
    const alert = document.createElement('div');
    alert.className = `alert ${alertClass} alert-dismissible fade show position-fixed`;
    alert.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
    alert.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(alert);
    
    // Auto-remove after 3 seconds
    setTimeout(function() {
        if (alert.parentNode) {
            alert.remove();
        }
    }, 3000);
}

function downloadCSV(csv, filename) {
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
}

// Vote confirmation
function confirmVote() {
    const selectedCandidate = document.querySelector('input[name="candidate_id"]:checked');
    if (!selectedCandidate) {
        showNotification('Please select a candidate before submitting your vote.', 'warning');
        return false;
    }
    
    const candidateName = selectedCandidate.nextElementSibling.querySelector('.card-title, h6, h5').textContent;
    return confirm(`Are you sure you want to vote for "${candidateName}"?\n\nThis action cannot be undone.`);
}

// Loading states
function showLoading(element) {
    element.classList.add('loading');
    const originalText = element.innerHTML;
    element.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>Loading...';
    element.disabled = true;
    
    return function hideLoading() {
        element.classList.remove('loading');
        element.innerHTML = originalText;
        element.disabled = false;
    };
}

// Form submission with loading
document.querySelectorAll('form').forEach(function(form) {
    form.addEventListener('submit', function(e) {
        const submitBtn = form.querySelector('button[type="submit"]');
        if (submitBtn && !submitBtn.classList.contains('no-loading')) {
            showLoading(submitBtn);
        }
    });
});

// Auto-refresh for active elections (every 30 seconds)
if (window.location.pathname.includes('/reports') || window.location.pathname.includes('/votes')) {
    setInterval(function() {
        // Only refresh if the page is visible
        if (!document.hidden) {
            window.location.reload();
        }
    }, 30000);
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Ctrl+Enter to submit forms
    if (e.ctrlKey && e.key === 'Enter') {
        const form = document.querySelector('form');
        if (form) {
            form.submit();
        }
    }
    
    // Escape to close modals
    if (e.key === 'Escape') {
        const modals = document.querySelectorAll('.modal.show');
        modals.forEach(function(modal) {
            bootstrap.Modal.getInstance(modal).hide();
        });
    }
});

// Print functionality
window.printReport = function() {
    window.print();
};

// Search functionality for tables
function initializeTableSearch() {
    const searchInput = document.getElementById('tableSearch');
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            const filter = this.value.toLowerCase();
            const rows = document.querySelectorAll('tbody tr');
            
            rows.forEach(function(row) {
                const text = row.textContent.toLowerCase();
                row.style.display = text.includes(filter) ? '' : 'none';
            });
        });
    }
}

// Initialize search on page load
initializeTableSearch();
