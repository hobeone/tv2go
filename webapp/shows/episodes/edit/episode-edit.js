angular.module('shows.episodes.edit',[])
.config(function($stateProvider){
  $stateProvider
  .state('tv2go.shows.episodes.edit', {
    url: "/episodes/:episodeId/edit/",
    templateUrl: "shows/episodes/edit/episode-edit.tmpl.html",
    controller: "EditEpisodeCtrl as editEpisodeCtrl"
  });
})
.controller('EditEpisodeCtrl', function($state, $stateParams, EpisodesModel){
  var editEpisodeCrtl = this;

  function returnToEpisodes() {
    $state.go("tv2go.shows.episodes", {
      show: $stateParams.show
    });
  }

  function cancelEditing() {
    returnToEpisodes();
  }

  EpisodesModel.getEpisodeById($stateParams.episodeId)
  .then(function(episode){
    if(episode) {
      editEpisodeCrtl.episode = episode;
      editEpisodeCrtl.editedEpisode = angular.copy(editEpisodeCrtl.episode);
    } else {
      returnToEpisodes();
    }
  });

  function updateEpisode() {
    editEpisodeCrtl.episode = angular.copy(editEpisodeCrtl.editedEpisode);
    EpisodesModel.updateEpisode(editEpisodeCrtl.episode);
    returnToEpisodes();
  }

  editEpisodeCrtl.cancelEditing = cancelEditing;
  editEpisodeCrtl.updateEpisode = updateEpisode;
})
;
