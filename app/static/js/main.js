document.querySelectorAll('.card').forEach((el, idx) => {
  el.classList.add('reveal');
  el.style.animationDelay = `${Math.min(idx * 60, 360)}ms`;
});

document.querySelectorAll('#toast-root .toast, main.container > .toast').forEach((toast) => {
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateY(8px)';
    setTimeout(() => toast.remove(), 220);
  }, 3600);
});

document.addEventListener('click', (event) => {
  const disabledButton = event.target.closest('button:disabled');
  if (disabledButton) event.preventDefault();
});
