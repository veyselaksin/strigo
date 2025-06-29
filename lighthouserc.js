module.exports = {
  ci: {
    collect: {
      staticDistDir: "./docs/_site",
      url: [
        "http://localhost/strigo/",
        "http://localhost/strigo/getting-started",
        "http://localhost/strigo/api",
        "http://localhost/strigo/advanced",
      ],
    },
    assert: {
      assertions: {
        "categories:performance": ["warn", { minScore: 0.8 }],
        "categories:accessibility": ["warn", { minScore: 0.9 }],
        "categories:best-practices": ["warn", { minScore: 0.8 }],
        "categories:seo": ["warn", { minScore: 0.8 }],
      },
    },
    upload: {
      target: "temporary-public-storage",
    },
  },
};
