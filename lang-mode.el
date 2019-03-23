(defconst lang-keywords
  (list
   "var"
   "func"
   "if"
   "else"
   "while"
   "for"
   "in"
   "continue"
   "break"
   "return"))

(defconst lang-types
  (list
   "int"
   "bool"
   "string"
   "range"))

(defconst lang-builtins
  (list
   "true"
   "false"
   "puts"
   "printf"))

(defconst lang-font-lock-keywords-1
  `(;; Keywords
    (,(rx symbol-start
          (eval `(or ,@lang-keywords))
          symbol-end)
     0 font-lock-keyword-face)

    ;; Types
    (,(rx symbol-start
          (eval `(or ,@lang-types))
          symbol-end)
     0 font-lock-type-face)

    ;; Builtins
    (,(rx symbol-start
          (eval `(or ,@lang-builtins))
          symbol-end)
     0 font-lock-builtin-face)

    ;; Function call
    (,(rx symbol-start
          (group (1+ (or alnum "_")))
          (0+ space)
          "(")
     1 font-lock-function-name-face)))

(defvar lang-mode-syntax-table
  (let ((tab (make-syntax-table text-mode-syntax-table)))
    (modify-syntax-entry ?\# "<" tab)
    (modify-syntax-entry ?\n ">" tab)
    (modify-syntax-entry ?\" "\"\"" tab)
    (modify-syntax-entry ?\' "\"'" tab)
    (modify-syntax-entry ?\\ "\\" tab)
    (modify-syntax-entry ?$ "'" tab)
    tab)
  "Syntax table for `lang-mode'.")

(define-derived-mode lang-mode prog-mode "lang"
  "Major mode for editing lang files."
  :syntax-table lang-mode-syntax-table
  (setq-local font-lock-defaults '(lang-font-lock-keywords-1))
  (setq-local comment-start "# ")
  (setq-local comment-start-skip "#+[\t ]*"))

(add-to-list 'auto-mode-alist '("\\.lg\\'" . lang-mode))

(provide 'lang-mode)
